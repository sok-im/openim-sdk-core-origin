package api

// LinkPreviewReq matches HTTP POST /link/preview.
type LinkPreviewReq struct {
	URL string `json:"url"`
}

// LinkPreviewResp is the structured Open Graph / meta preview for a URL.
type LinkPreviewResp struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	SiteName    string `json:"siteName"`
}

var (
	LinkPreview = newApi[LinkPreviewReq, LinkPreviewResp]("/link/preview")
)
