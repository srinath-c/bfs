package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Page struct {
	Title     string
	FileInfos []*Info
}

type Info struct {
	IsDir     bool
	Href      string
	Filename  string
	Extension string
}

type RangeReader struct {
	start, end int64
	read       int64
	seeked     bool
	rd         *os.File
}

var base = flag.String("dir", ".", "base directory")
var addr = flag.String("addr", ":8080", "address:port to listen at")

func (rr *RangeReader) Read(p []byte) (int, error) {
	fmt.Println("Start: ", rr.start, "End: ", rr.end)
	if !rr.seeked {
		rr.seeked = true
		if rr.start < 0 {
			_, err := rr.rd.Seek(rr.end+rr.start, 0)
			if err != nil {
				return 0, err
			}
		} else {
			_, err := rr.rd.Seek(rr.start, 0)
			if err != nil {
				return 0, err
			}
		}
	}

	var r int
	var err error

	capacity := cap(p)
	if int64(capacity) > rr.end-rr.start+1-rr.read {
		temp := make([]byte, rr.end-rr.start+1-rr.read)
		r, err = rr.rd.Read(temp)
		if err != nil {
			return r, err
		}
		copy(p, temp)
		return r, io.EOF
	}
	r, err = rr.rd.Read(p)
	if err != nil {
		return r, err
	}

	rr.read += int64(r)

	return r, nil
}

func serveFile(w http.ResponseWriter, r *http.Request, filename string) {
	w.Header().Set("Accept-Ranges", "bytes")
	fh, err := os.Open(filename)
	if err != nil {
		w.Header().Set("BFS_Error", err.Error())
		w.WriteHeader(502)
		w.Write([]byte("Sorry dude...Workin on it..."))
		return
	}
	defer fh.Close()

	stat, err := fh.Stat()
	if err != nil {
		w.Header().Set("BFS_Error", err.Error())
		w.WriteHeader(502)
		w.Write([]byte("Unable stat after opening a file....Crazy right!!"))
		return
	}

	start, end, err := parseRangeHeader(r.Header.Get("range"))
	if err != nil {
		w.Header().Set("BFS_Error", err.Error())
		w.WriteHeader(416)
		w.Write([]byte("Range is not valid"))
		return
	}
	if (end > 0 && start > end) || start >= stat.Size() || end > stat.Size() {
		w.Header().Set("BFS_Error", err.Error())
		w.WriteHeader(416)
		w.Write([]byte("Range is not valid"))
		return
	}

	var h io.Reader
	if start == 0 && end < 0 {
		end = stat.Size()
		h = fh
	} else {
		if end < 0 {
			end = stat.Size()
		}
		h = &RangeReader{start: start, end: end, rd: fh}
	}
	cl := int64(0)
	if start < 0 {
		cl = start
	} else {
		cl = end - start
	}
	w.Header().Set("Content-Length", fmt.Sprint(cl))
	io.Copy(w, h)
	return
}

func handle(w http.ResponseWriter, r *http.Request) {
	isDir := false
	if strings.HasSuffix(r.URL.Path, "/") {
		isDir = true
	}
	clean := filepath.Clean(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/"))
	file := filepath.Join(*base, clean)

	stat, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(404)
			w.Write([]byte("File not found"))
			return
		}
		w.Header().Set("BFS_Error", err.Error())
		w.WriteHeader(502)
		w.Write([]byte("Sorry dude...Workin on it..."))
		return
	}

	if !stat.IsDir() {
		if isDir {
			w.WriteHeader(404)
			w.Write([]byte("Directory not found"))
			return
		}
		http.ServeFile(w, r, file)
		return
	}
	dir, err := ioutil.ReadDir(file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(502)
		w.Write([]byte("Fatal Server error...You are being blamed for the shit!"))
		return
	}

	w.Header().Set("Content-Type", "text/html")

	p := &Page{}
	p.Title = "Index of " + filepath.Base(file)
	p.FileInfos = make([]*Info, 0)
	for _, finfo := range dir {
		p.FileInfos = append(p.FileInfos, &Info{
			IsDir:     finfo.IsDir(),
			Href:      filepath.Join("/", clean, finfo.Name()),
			Filename:  finfo.Name(),
			Extension: mime.TypeByExtension(filepath.Ext(finfo.Name())),
		})
	}
	t := template.New("base")
	t, err = t.Parse(html)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(502)
		w.Write([]byte("Template parse error"))
		return
	}
	err = t.Execute(w, p)
	if err != nil {
		log.Println(err.Error())
	}
}

func parseRangeHeader(header string) (int64, int64, error) {
	if header == "" {
		return 0, -1, nil
	}

	if len(header) < 6 {
		return 0, 0, errors.New("invalid range")
	}

	if header[:6] != "bytes=" {
		return 0, 0, errors.New("only byte ranges suppported")
	}

	if header[6:] == "" {
		return 0, -1, nil
	}

	if len(header[6:]) < 2 {
		return 0, 0, errors.New("invalid range")
	}

	var start, end int64

	if header[6:][0] == '-' {
		_, err := fmt.Sscanf(header[6:], "%d", &start)
		if err != nil {
			return 0, 0, err
		}
		return start, -1, nil
	}

	if header[6:][len(header[6:])-1] == '-' {
		_, err := fmt.Sscanf(header[6:], "%d-", &start)
		if err != nil {
			return 0, 0, err
		}
		return start, -1, nil
	}

	_, err := fmt.Sscanf(header[6:], "%d-%d", &start, &end)
	if err != nil {
		return 0, 0, err
	}

	return start, end, nil
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handle)
	fmt.Println("Starting bfs on ", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
