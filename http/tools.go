package http

import (
	"fmt"
	fbErrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/img"
	"net/http"
	"net/url"
)

var toolsPatchHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	src := r.URL.Path
	action := r.URL.Query().Get("action")
	dst := r.URL.Query().Get("destination")
	dst, err := url.QueryUnescape(dst)
	fps := r.URL.Query().Get("fps")
	src = d.user.FullPath(src)
	switch action {
	case "copyExif":
		dst = d.user.FullPath(dst)
		err = img.CopyExif(src, dst)
	case "extractFrame":
		err = img.ExtractFrame(src, fps)
	case "extractExif":
		var exifData img.Exif
		err = img.ExtractExif(src, &exifData)
		if err == nil {
			return renderJSON(w, r, exifData)
		}
	default:
		err = fmt.Errorf("unsupported action %s: %w", action, fbErrors.ErrInvalidRequestParams)
	}
	return errToStatus(err), err
})
