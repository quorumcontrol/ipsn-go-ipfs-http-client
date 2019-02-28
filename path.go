package httpapi

import (
	"context"

	cid "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-cid"
	ipld "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipld-format"
	ipfspath "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-path"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/interface-go-ipfs-core"
)

func (api *HttpApi) ResolvePath(ctx context.Context, path iface.Path) (iface.ResolvedPath, error) {
	var out struct {
		Cid     cid.Cid
		RemPath string
	}

	//TODO: this is hacky, fixing https://github.com/ipfs/go-ipfs/issues/5703 would help

	var err error
	if path.Namespace() == "ipns" {
		if path, err = api.Name().Resolve(ctx, path.String()); err != nil {
			return nil, err
		}
	}

	if err := api.request("dag/resolve", path.String()).Exec(ctx, &out); err != nil {
		return nil, err
	}

	// TODO:
	ipath, err := ipfspath.FromSegments("/"+path.Namespace()+"/", out.Cid.String(), out.RemPath)
	if err != nil {
		return nil, err
	}

	root, err := cid.Parse(ipfspath.Path(path.String()).Segments()[1])
	if err != nil {
		return nil, err
	}

	return iface.NewResolvedPath(ipath, out.Cid, root, out.RemPath), nil
}

func (api *HttpApi) ResolveNode(ctx context.Context, p iface.Path) (ipld.Node, error) {
	rp, err := api.ResolvePath(ctx, p)
	if err != nil {
		return nil, err
	}

	return api.Dag().Get(ctx, rp.Cid())
}
