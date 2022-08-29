package server

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/luoruofeng/httpchunktransfer/common"
	cnf "github.com/luoruofeng/httpchunktransfer/config"
)

func info(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	fi, err := os.Stat(cnf.C.Filesys.Rootpath + filename)
	if err != nil && os.IsNotExist(err) {
		log.Println(err)
		fmt.Fprintln(w, "file:", filename, " is not exist in root path:", cnf.C.Filesys.Rootpath)
		return
	}
	f, _ := os.Open(cnf.C.Filesys.Rootpath + filename)
	defer f.Close()
	w.Header().Set("My-File-Length", strconv.FormatInt(fi.Size(), 10))
	w.Header().Add("Accept-Ranges", "bytes") // indicate support range request. if not support set response header Accept-Ranges:none
}

// http range
func hrange(w http.ResponseWriter, r *http.Request) {
	//get http request header
	rv := r.Header.Get("Range")
	id := r.Header.Get("My-Chunk-Id")
	// ae := r.Header.Get("Accept-Encoding")

	//get first and second number in http request header 'Range'
	vaild := regexp.MustCompile("[0-9]+")
	ss, es := vaild.FindAllString(rv, -1)[0], vaild.FindAllString(rv, -1)[1]
	s, _ := strconv.ParseInt(ss, 10, 64) // start index of chunck part in total file
	e, _ := strconv.ParseInt(es, 10, 64) // to end index of chunck part in total file(chunked part include this number)

	//check file is exist
	downloadpath := r.URL.Query().Get("filepath")
	if filestat, err := os.Stat(downloadpath); err != nil || os.IsNotExist(err) {
		fmt.Println("download file is not exist!")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "download file is not exist!")
		return
	} else {
		filesize := filestat.Size()
		//check http request header end range is less than file size
		if e >= filesize {
			io.WriteString(w, fmt.Sprintf("Http request header Range is greater than file size! range is %v-%v. file size is:%v", s, e, filesize))
			return
		}

		f, _ := os.Open(downloadpath)
		defer f.Close()

		b := make([]byte, cnf.C.Filesys.Blocksize)  // for multiple part of chuncked file send to client t in batches
		ab := make([]byte, cnf.C.Filesys.Chunksize) // all chunk byte array for get md5 value
		var ti int64 = s                            // record start index of chunck part in total file
		var i int64 = 0                             // record current index of below loop
		for {
			rn, err := f.ReadAt(b, ti) //read a part of chunck to b
			n := int64(rn)
			if n <= 0 && err == io.EOF {
				fmt.Println("finish!")
				break
			}
			var ib []byte
			if ti+n-1 >= e { // if start index of chunck plus cursor index sub 1 greater than or equal to end index of chunck
				fmt.Printf("id:%v e:%v index:%v this chunck read finish\n", id, e, ti)
				ib = b[:e-ti+1] // ib is the last block read  of chunck  minus the excess
				copy(ab[i:], ib)
				ab = ab[:(i + (e - ti + 1))]
				break
			} else {
				fmt.Printf("id:%v e:%v index:%v\n", id, e, ti)
				ib = b[:n]
				//because of md5 should be send by http response header and set header have to before response write.
			} // we could not response write at this time.
			copy(ab[i:], ib)
			ti += n
			i += n
		}
		w.Header().Add("My-File-MD5", fmt.Sprintf("%x", common.GetBytesMd5(ab)))
		w.Header().Add("Transfer-Encoding", "chunked") //
		w.Header().Add("Accept-Ranges", "bytes")       // indicate support range request. if not support set response header Accept-Ranges:none
		w.Header().Add("Content-Range", fmt.Sprintf("bytes %v-%v/%v", s, e, filesize))
		if cnf.C.Filesys.Compress {
			w.Header().Add("Content-Encoding", "gzip") // compress by gzip
		}
		w.WriteHeader(http.StatusPartialContent) // Must put this line behind set header operation otherwise set header operation won't work. set http status code equal 206,indicate partial content

		if cnf.C.Filesys.Compress {
			//compress
			gw := gzip.NewWriter(w)
			gw.Flush()
			defer gw.Close()
			io.Copy(gw, bytes.NewReader(ab))
		} else {
			io.Copy(w, bytes.NewReader(ab))
		}
	}
}

func checkChunck(target []byte, destination []byte) bool {
	tencoded := base64.StdEncoding.EncodeToString([]byte(target))
	dencoded := base64.StdEncoding.EncodeToString([]byte(destination))
	return tencoded == dencoded
}
