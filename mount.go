package nfs

import (
	"bytes"
	"context"

	"github.com/willscott/go-nfs-client/nfs/xdr"
)

const (
	mountServiceID = 100005
)

func init() {
	_ = RegisterMessageHandler(mountServiceID, uint32(MountProcNull), onMountNull)
	_ = RegisterMessageHandler(mountServiceID, uint32(MountProcMount), onMount)
	_ = RegisterMessageHandler(mountServiceID, uint32(MountProcUmnt), onUMount)
	_ = RegisterMessageHandler(mountServiceID, uint32(MountProcExport), onExport)
}

func onMountNull(ctx context.Context, w *response, userHandle Handler) error {
	return w.writeHeader(ResponseCodeSuccess)
}

func onMount(ctx context.Context, w *response, userHandle Handler) error {
	// TODO: auth check.
	dirpath, err := xdr.ReadOpaque(w.req.Body)
	if err != nil {
		return err
	}
	mountReq := MountRequest{Header: w.req.Header, Dirpath: dirpath}
	status, handle, flavors := userHandle.Mount(ctx, w.conn, mountReq)

	if err := w.writeHeader(ResponseCodeSuccess); err != nil {
		return err
	}

	writer := bytes.NewBuffer([]byte{})
	if err := xdr.Write(writer, uint32(status)); err != nil {
		return err
	}

	rootHndl := userHandle.ToHandle(handle, []string{})

	if status == MountStatusOk {
		_ = xdr.Write(writer, rootHndl)
		_ = xdr.Write(writer, flavors)
	}
	return w.Write(writer.Bytes())
}

func onUMount(ctx context.Context, w *response, userHandle Handler) error {
	_, err := xdr.ReadOpaque(w.req.Body)
	if err != nil {
		return err
	}

	return w.writeHeader(ResponseCodeSuccess)
}

// onExport returns the list of exported filesystems (RFC 1813, appendix I,
// MOUNTPROC3_EXPORT). We export a single root "/", with no host restrictions.
func onExport(ctx context.Context, w *response, userHandle Handler) error {
	if err := w.writeHeader(ResponseCodeSuccess); err != nil {
		return err
	}

	writer := bytes.NewBuffer([]byte{})
	_ = xdr.Write(writer, uint32(1))        // value_follows: one export entry
	_ = xdr.Write(writer, []byte("/"))       // ex_dir
	_ = xdr.Write(writer, uint32(0))         // ex_groups: none (no host restrictions)
	_ = xdr.Write(writer, uint32(0))         // ex_next: end of list
	return w.Write(writer.Bytes())
}
