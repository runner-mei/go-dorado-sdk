package dorado

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type LunCopy struct {
	BASELUN               string `json:"BASELUN"`
	COPYPROGRESS          string `json:"COPYPROGRESS"`
	COPYSPEED             string `json:"COPYSPEED"`
	COPYSTARTTIME         string `json:"COPYSTARTTIME"`
	COPYSTOPTIME          string `json:"COPYSTOPTIME"`
	DESCRIPTION           string `json:"DESCRIPTION"`
	HEALTHSTATUS          string `json:"HEALTHSTATUS"`
	ID                    int    `json:"ID,string"`
	LUNCOPYTYPE           string `json:"LUNCOPYTYPE"`
	NAME                  string `json:"NAME"`
	RUNNINGSTATUS         string `json:"RUNNINGSTATUS"`
	SOURCELUN             string `json:"SOURCELUN"`
	SOURCELUNCAPACITY     string `json:"SOURCELUNCAPACITY"`
	SOURCELUNCAPACITYBYTE string `json:"SOURCELUNCAPACITYBYTE"`
	SOURCELUNNAME         string `json:"SOURCELUNNAME"`
	SOURCELUNWWN          string `json:"SOURCELUNWWN"`
	SUBTYPE               string `json:"SUBTYPE"`
	TARGETLUN             string `json:"TARGETLUN"`
	TYPE                  int    `json:"TYPE"`
}

const (
	DefaultLUNCopyTimeoutSecond = 60
)

func (d *Device) GetLUNCopys(ctx context.Context, query *SearchQuery) ([]LunCopy, error) {
	spath := "/luncopy"

	req, err := d.newRequest(ctx, "GET", spath, nil)
	if err != nil {
		return nil, fmt.Errorf(ErrCreateRequest+": %w", err)
	}

	req = AddSearchQuery(req, query)

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(ErrHTTPRequestDo+": %w", err)
	}

	lunCopys := []LunCopy{}
	if err = decodeBody(resp, &lunCopys); err != nil {
		return nil, fmt.Errorf(ErrDecodeBody+": %w", err)
	}

	if len(lunCopys) == 0 {
		return nil, errors.New(ErrLunNotFound)
	}

	return lunCopys, nil
}

func (d *Device) GetLUNCopy(ctx context.Context, lunCopyID int) (*LunCopy, error) {
	spath := fmt.Sprintf("/luncopy/%d", lunCopyID)

	req, err := d.newRequest(ctx, "GET", spath, nil)
	if err != nil {
		return nil, fmt.Errorf(ErrCreateRequest+": %w", err)
	}
	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(ErrHTTPRequestDo+": %w", err)
	}

	lunCopys := &LunCopy{}
	if err = decodeBody(resp, lunCopys); err != nil {
		return nil, fmt.Errorf(ErrDecodeBody+": %w", err)
	}

	return lunCopys, nil
}

func (d *Device) CreateLUNCopy(ctx context.Context, sourceLUNID, targetLUNID int) (*LunCopy, error) {
	spath := "/luncopy"
	param := struct {
		NAME      string `json:"NAME"`
		SOURCELUN string `json:"SOURCELUN"`
		TARGETLUN string `json:"TARGETLUN"`
		COPYSPEED int    `json:"COPYSPEED"`
	}{
		NAME:      fmt.Sprintf("LUNCopy_%d_%d", sourceLUNID, targetLUNID),
		SOURCELUN: fmt.Sprintf("INVALID;%d;INVALID;INVALID;INVALID", sourceLUNID),
		TARGETLUN: fmt.Sprintf("INVALID;%d;INVALID;INVALID;INVALID", targetLUNID),
		COPYSPEED: 4,
	}
	jb, err := json.Marshal(param)
	if err != nil {
		return nil, fmt.Errorf(ErrCreatePostValue+": %w", err)
	}

	req, err := d.newRequest(ctx, "POST", spath, bytes.NewBuffer(jb))
	if err != nil {
		return nil, fmt.Errorf(ErrCreateRequest+": %w", err)
	}
	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(ErrHTTPRequestDo+": %w", err)
	}

	luncopy := &LunCopy{}
	if err = decodeBody(resp, luncopy); err != nil {
		return nil, fmt.Errorf(ErrDecodeBody+": %w", err)
	}

	return luncopy, nil
}

func (d *Device) DeleteLUNCopy(ctx context.Context, luncopyID int) error {
	spath := fmt.Sprintf("/luncopy/%d", luncopyID)

	req, err := d.newRequest(ctx, "DELETE", spath, nil)
	if err != nil {
		return fmt.Errorf(ErrCreateRequest+": %w", err)
	}
	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(ErrHTTPRequestDo+": %w", err)
	}

	var i interface{} // this endpoint return N/A
	if err = decodeBody(resp, i); err != nil {
		return fmt.Errorf(ErrDecodeBody+": %w", err)
	}

	return nil
}

func (d *Device) StartLUNCopy(ctx context.Context, luncopyID int) error {
	spath := "/luncopy/start"
	param := struct {
		TYPE string `json:"TYPE"`
		ID   string `json:"ID"`
	}{
		TYPE: strconv.Itoa(TypeLUNCopy),
		ID:   strconv.Itoa(luncopyID),
	}
	jb, err := json.Marshal(param)
	if err != nil {
		return fmt.Errorf(ErrCreatePostValue+": %w", err)
	}

	req, err := d.newRequest(ctx, "PUT", spath, bytes.NewBuffer(jb))
	if err != nil {
		return fmt.Errorf(ErrCreateRequest+": %w", err)
	}
	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(ErrHTTPRequestDo+": %w", err)
	}

	var i interface{} // this endpoint return N/A
	if err = decodeBody(resp, i); err != nil {
		return fmt.Errorf(ErrDecodeBody+": %w", err)
	}

	return nil
}

// CreateLUNCopyWithWait start luncopy and wait to copy
func (d *Device) StartLUNCopyWithWait(ctx context.Context, luncopyID int, timeoutCount int) error {
	if timeoutCount == 0 {
		timeoutCount = DefaultLUNCopyTimeoutSecond
	}

	err := d.StartLUNCopy(ctx, luncopyID)
	if err != nil {
		return fmt.Errorf("failed to start luncopy (ID: %d): %w", luncopyID, err)
	}

	// wait 60 seconds (default)
	for i := 0; i < timeoutCount; i++ {
		isReady, err := d.luncopyIsDone(ctx, luncopyID)
		if err != nil {
			return fmt.Errorf("failed to wait that luncopy is done: %w", err)
		}

		if isReady == true {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *Device) luncopyIsDone(ctx context.Context, luncopyID int) (bool, error) {
	luncopy, err := d.GetLUNCopy(ctx, luncopyID)
	if err != nil {
		return false, fmt.Errorf("failed to get luncopy (ID: %d): %w", luncopyID, err)
	}

	if luncopy.HEALTHSTATUS != strconv.Itoa(StatusHealth) {
		return false, fmt.Errorf("luncopy health status is bad (HEALTHSTATUS: %s)", luncopy.HEALTHSTATUS)
	}

	if luncopy.RUNNINGSTATUS == strconv.Itoa(StatusLunCopyReady) {
		return true, nil
	}

	return false, nil
}