package client

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/luoruofeng/httpchunktransfer/common"
	cnf "github.com/luoruofeng/httpchunktransfer/config"
)

//no support http range error
type NotRangeRequestError struct {
	StatusCode int
	Err        error
}

func (e *NotRangeRequestError) Error() string {
	return e.Err.Error()
}

func (e *NotRangeRequestError) Temporary() bool {
	return e.StatusCode == http.StatusServiceUnavailable // 503
}

func queryInfo(filename string) (*http.Response, error) {
	//get file size
	req, _ := http.NewRequest("HEAD", common.GetAddr()+"/info?filename="+filename, nil)
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	resp.Body.Close()

	if resp.Header.Get("Accept-Ranges") == "" || resp.Header.Get("Accept-Ranges") != "bytes" {
		return nil, &NotRangeRequestError{
			StatusCode: 503,
			Err:        errors.New("can not support http range request!"),
		}
	}
	return resp, nil
}

func createSubFile(filename string, id int) (sf *os.File) {
	fsd := strings.Split(filename, ".")
	sf, _ = os.Create(filepath.Join(cnf.C.Filesys.Savedir, fmt.Sprintf("%v%02d.%v", fsd[0], id, fsd[1])))
	return sf
}

// request chunk of file by http
//cs is chunk size
//fs is file total size
//nc is number of chunk
func requestHttpChunk(id int, cs uint64, fs uint64, nc int, filename string) {
	si := cs * uint64(id) // start index of chunck
	ei := si + cs - 1     // end index of chunck
	if id == nc-1 {       // last chunk need to change ei
		ei = fs - 1
	}

	//request url with request header:
	//e.x Content-Range: bytes 0-38615451/38615452
	//0 is the first character
	//38615451 is the last charactor
	//38615452 is file size(byte unit)
	//example above only have one chunked
	req, _ := http.NewRequest("GET", common.GetAddr()+"/hrange", nil)
	values := req.URL.Query()
	values.Add("filepath", filepath.Join(cnf.C.Filesys.Rootpath, filename))
	req.URL.RawQuery = values.Encode()
	req.Header.Add("Range", fmt.Sprintf("bytes=%v-%v", si, ei))
	req.Header.Add("My-Chunk-Id", strconv.Itoa(id))
	req.Header.Add("Accept-Encoding", "gzip")
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	//check response header attribute Content-Encoding value is gzip
	var gr io.ReadCloser
	var er error
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gr, er = gzip.NewReader(resp.Body) // need compress
		if er != nil {
			fmt.Println("id:", id, er)
			os.Exit(1)
		}
		defer gr.Close()
	default:
		gr = resp.Body // no need decompress
	}

	//check chunk md5 of recive file
	rmd5 := resp.Header.Get("My-File-MD5") // response header return chunked MD5
	bb, _ := io.ReadAll(gr)                // body byte array of response
	if rmd5 == fmt.Sprintf("%x", common.GetBytesMd5(bb)) {
		fmt.Println("id:" + strconv.Itoa(id) + " client: chunked receive successfully")
		rb := ioutil.NopCloser(bytes.NewReader(bb))
		defer rb.Close()
		// create sub file
		sf := createSubFile(filename, id)
		defer sf.Close()
		io.Copy(sf, rb) // write to new chunked file
	} else {
		fmt.Println("id:" + strconv.Itoa(id) + " client: chunked receive unsuccessfully")
	}
}

func Query(filename string) {
	if resp, err := queryInfo(filename); err != nil { //get file info
		log.Println(err)
		return
	} else {
		//set up file arguments
		fss := resp.Header.Get("My-File-Length")        //get file size
		fs, _ := strconv.ParseUint(fss, 10, 64)         //file total size
		cs := uint64(cnf.C.Filesys.Chunksize)           //chunked size
		nc := int(math.Ceil(float64(fs) / float64(cs))) // number of chunked

		var wg sync.WaitGroup
		for i := 0; i < nc; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				requestHttpChunk(id, cs, fs, nc, filename)
			}(i)
		}
		wg.Wait()

		//combine sub file and delete sub file
		combine(filename)

		//for test
		if checkChunck(common.GetFileMd5(filepath.Join(cnf.C.Filesys.Rootpath, filename)), common.GetFileMd5(filepath.Join(cnf.C.Filesys.Savedir, filename))) {
			fmt.Println("completeness of file has checked successfully!")
		} else {
			fmt.Println("completeness of file error!")
		}
	}

}

func combine(newfilename string) {
	// sort children files of savedir by name
	dir, _ := os.Open(cnf.C.Filesys.Savedir)
	defer dir.Close()
	files, _ := dir.Readdir(-1)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	//create combine file
	nf, _ := os.Create(filepath.Join(cnf.C.Filesys.Savedir, newfilename))
	defer nf.Close()

	for _, fi := range files {
		if fi.Name() == newfilename {
			continue
		}
		f, _ := os.Open(filepath.Join(cnf.C.Filesys.Savedir, fi.Name()))
		io.Copy(nf, f)
		f.Close()
		fmt.Println("combine subfile:", f.Name(), " to:", filepath.Join(cnf.C.Filesys.Savedir, newfilename))
	}

	fmt.Println("Combine Done!!")

	for _, fi := range files {
		if fi.Name() == newfilename {
			continue
		}
		os.Remove(filepath.Join(cnf.C.Filesys.Savedir, fi.Name()))
		fmt.Println("remove chunk file:", fi.Name())
	}
}

func checkChunck(target []byte, destination []byte) bool {
	tencoded := base64.StdEncoding.EncodeToString([]byte(target))
	dencoded := base64.StdEncoding.EncodeToString([]byte(destination))
	return tencoded == dencoded
}
