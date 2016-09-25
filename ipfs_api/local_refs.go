package ipfs_api

import (
	"encoding/json"

	"github.com/ipfs/go-ipfs-api"
)

func (s *MyShell) LocalRefs() (<-chan string, error) {
	req := shell.NewRequest(s.Url, "refs/local")

	resp, err := req.Send(s.Client)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	out := make(chan string)
	go func() {
		var ref struct {
			Ref string
		}
		defer resp.Close()
		defer close(out)
		dec := json.NewDecoder(resp.Output)
		for {
			err := dec.Decode(&ref)
			if err != nil {
				return
			}
			if len(ref.Ref) > 0 {
				out <- ref.Ref
			}
		}
	}()

	return out, nil
}
