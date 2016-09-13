package ipfs_api

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/ipfs/go-ipfs-api"
	files "github.com/whyrusleeping/go-multipart-files"
)

type object struct {
	Hash string
}

// Add a file to ipfs from the given reader, returns the hash of the added file
func (s *MyShell) Add(r io.Reader) (string, error) {
	var rc io.ReadCloser
	if rclose, ok := r.(io.ReadCloser); ok {
		rc = rclose
	} else {
		rc = ioutil.NopCloser(r)
	}

	// handler expects an array of files
	fr := files.NewReaderFile("", "", rc, nil)
	slf := files.NewSliceFile("", "", []files.File{fr})
	fileReader := files.NewMultiFileReader(slf, true)

	req := shell.NewRequest(s.Url, "add")
	req.Body = fileReader
	req.Opts["progress"] = "false"
	req.Opts["chunker"] = "size-1048576"

	resp, err := req.Send(s.Client)
	if err != nil {
		return "", err
	}
	defer resp.Close()
	if resp.Error != nil {
		return "", resp.Error
	}

	var out object
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Hash, nil
}
