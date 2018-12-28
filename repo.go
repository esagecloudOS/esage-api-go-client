package abiquo_api

import (
	"encoding/json"
	// "log"
)

type RepoCollection struct {
	AbstractCollection
	Collection []Repo
}

type Repo struct {
	DTO
	Name               string
	RepositoryLocation string
}

func (r *Repo) GetTemplates(c *AbiquoClient) ([]VirtualMachineTemplate, error) {
	var templatesCol TemplateCollection
	var templates []VirtualMachineTemplate

	templates_resp, err := r.FollowLink("virtualmachinetemplates", c)
	if err != nil {
		return templates, err
	}
	json.Unmarshal(templates_resp.Body(), &templatesCol)

	for {
		for _, t := range templatesCol.Collection {
			// l, _ := t.GetLink("edit")
			// log.Printf("CLIENT REPO == Found template %s at %s", t.Name, l.Href)
			templates = append(templates, t)
		}
		if templatesCol.HasNext() {
			next_link := templatesCol.GetNext()
			templates_raw, err := c.checkResponse(c.client.R().SetHeader("Accept", "application/vnd.abiquo.virtualmachinetemplates+json").
				Get(next_link.Href))
			if err != nil {
				return templates, err
			}
			templatesCol = TemplateCollection{}
			json.Unmarshal(templates_raw.Body(), &templatesCol)
		} else {
			break
		}
	}

	return templates, nil
}
