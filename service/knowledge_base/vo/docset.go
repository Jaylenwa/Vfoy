package vo

// DocSetCreateReq
type (
	DocSetCreateReq struct {
		Name      string     `json:"name"`
		Desc      string     `json:"desc"`
		Documents []Document `json:"documents"`
	}

	// Document 表示文档部分的信息
	Document struct {
		Name       string      `json:"name"`
		Paragraphs []Paragraph `json:"paragraphs"`
	}

	// Paragraph 表示段落信息
	Paragraph struct {
		Content     string    `json:"content"`
		Title       string    `json:"title"`
		IsActive    bool      `json:"is_active"`
		ProblemList []Problem `json:"problem_list"`
	}

	// Problem 表示问题列表中的元素
	Problem struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
)

// DocSetCreateRes
type (
	DocSetCreateRes struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    ResponseData `json:"data"`
	}

	ResponseData struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		Desc          string    `json:"desc"`
		UserID        string    `json:"user_id"`
		CharLength    int       `json:"char_length"`
		DocumentCount int       `json:"document_count"`
		UpdateTime    string    `json:"update_time"`
		CreateTime    string    `json:"create_time"`
		DocumentList  []DocInfo `json:"document_list"`
	}

	DocInfo struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		CharLength     int    `json:"char_length"`
		UserID         string `json:"user_id"`
		ParagraphCount int    `json:"paragraph_count"`
		IsActive       bool   `json:"is_active"`
		UpdateTime     string `json:"update_time"`
		CreateTime     string `json:"create_time"`
	}
)
