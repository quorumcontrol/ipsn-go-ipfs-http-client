package httpapi

import (
	"context"
	"errors"

	"github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/interface-go-ipfs-core"
	caopts "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-peer"
)

type KeyAPI HttpApi

type keyOutput struct {
	JName string `json:"Name"`
	Id    string

	pid peer.ID
}

func (k *keyOutput) Name() string {
	return k.JName
}

func (k *keyOutput) Path() iface.Path {
	p, _ := iface.ParsePath("/ipns/" + k.Id)
	return p
}

func (k *keyOutput) ID() peer.ID {
	return k.pid
}

func (api *KeyAPI) Generate(ctx context.Context, name string, opts ...caopts.KeyGenerateOption) (iface.Key, error) {
	options, err := caopts.KeyGenerateOptions(opts...)
	if err != nil {
		return nil, err
	}

	var out keyOutput
	err = api.core().request("key/gen", name).
		Option("type", options.Algorithm).
		Option("size", options.Size).
		Exec(ctx, &out)
	if err != nil {
		return nil, err
	}
	out.pid, err = peer.IDB58Decode(out.Id)
	return &out, err
}

func (api *KeyAPI) Rename(ctx context.Context, oldName string, newName string, opts ...caopts.KeyRenameOption) (iface.Key, bool, error) {
	options, err := caopts.KeyRenameOptions(opts...)
	if err != nil {
		return nil, false, err
	}

	var out struct {
		Was       string
		Now       string
		Id        string
		Overwrite bool
	}
	err = api.core().request("key/rename", oldName, newName).
		Option("force", options.Force).
		Exec(ctx, &out)
	if err != nil {
		return nil, false, err
	}

	id := &keyOutput{JName: out.Now, Id: out.Id}
	id.pid, err = peer.IDB58Decode(id.Id)
	return id, out.Overwrite, err
}

func (api *KeyAPI) List(ctx context.Context) ([]iface.Key, error) {
	var out struct{ Keys []*keyOutput }
	if err := api.core().request("key/list").Exec(ctx, &out); err != nil {
		return nil, err
	}

	res := make([]iface.Key, len(out.Keys))
	for i, k := range out.Keys {
		var err error
		k.pid, err = peer.IDB58Decode(k.Id)
		if err != nil {
			return nil, err
		}
		res[i] = k
	}

	return res, nil
}

func (api *KeyAPI) Self(ctx context.Context) (iface.Key, error) {
	var id struct{ ID string }
	if err := api.core().request("id").Exec(ctx, &id); err != nil {
		return nil, err
	}

	var err error
	out := keyOutput{JName: "self", Id: id.ID}
	out.pid, err = peer.IDB58Decode(out.Id)
	return &out, err
}

func (api *KeyAPI) Remove(ctx context.Context, name string) (iface.Key, error) {
	var out struct{ Keys []keyOutput }
	if err := api.core().request("key/rm", name).Exec(ctx, &out); err != nil {
		return nil, err
	}
	if len(out.Keys) != 1 {
		return nil, errors.New("got unexpected number of keys back")
	}

	var err error
	out.Keys[0].pid, err = peer.IDB58Decode(out.Keys[0].Id)
	return &out.Keys[0], err
}

func (api *KeyAPI) core() *HttpApi {
	return (*HttpApi)(api)
}
