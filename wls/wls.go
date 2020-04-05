package wls

type WLSRestQuery struct {
	Fields   []string                 `json:"fields"`
	Children map[string]*WLSRestQuery `json:"children,omitempty"`
	Links    []string                 `json:"links"`
}

type WLSRawResponse = map[string]interface{}
