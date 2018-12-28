package abiquo_api

import (
	"encoding/json"
)

type DiskCollection struct {
	AbstractCollection
	Collection []Disk
}

type Disk struct {
	DTO
	DiskController     string `json:"diskController,omitempty"`
	DiskControllerType string `json:"diskControllerType,omitempty"`
	Bus                int    `json:"bus,omitempty"`
	Unit               int    `json:"unit,omitempty"`
	Id                 int    `json:"id,omitempty"`
	Label              string `json:"label,omitempty"`
	Sequence           int    `json:"sequence,omitempty"`
	SizeInMb           int    `json:"sizeInMb,omitempty"`
	Uuid               string `json:"uuid,omitempty"`
	Allocation         string `json:"allocation,omitempty"`
	Path               string `json:"path,omitempty"`
	DiskFormatType     string `json:"diskFormatType,omitempty"`
	DiskFileSize       int    `json:"diskFileSize,omitempty"`
	HdRequired         int    `json:"hdRequired,omitempty"`
	State              string `json:"state,omitempty"`
	CreationDate       string `json:"creationDate,omitempty"`
	Bootable           bool   `json:"bootable,omitempty"`
}

func (d *Disk) Update(c *AbiquoClient) error {
	disk_lnk, _ := d.GetLink("edit")

	body_json, err := json.Marshal(d)
	if err != nil {
		return err
	}

	_, err = c.checkResponse(c.client.R().SetHeader("Accept", disk_lnk.Type).
		SetHeader("Content-Type", disk_lnk.Type).
		SetBody(body_json).
		Put(disk_lnk.Href))
	if err != nil {
		return err
	}
	return nil
}
