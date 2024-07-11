package entity

const (
	PROLOGUE = "您好，我是 MaxKB 小助手，您可以向我提出 MaxKB 使用问题。\n- MaxKB 主要功能有什么？\n- MaxKB 支持哪些大语言模型？\n- MaxKB 支持哪些文档类型？"
	PROMPT   = "已知信息：\n{data}\n回答要求：\n- 请使用简洁且专业的语言来回答用户的问题。\n- 如果你不知道答案，请回答“没有在知识库中查找到相关信息，建议咨询相关技术支持或参考官方文档进行操作”。\n- 避免提及你是从已知信息中获得的知识。\n- 请保证答案与已知信息中描述的一致。\n- 请使用 Markdown 语法优化答案的格式。\n- 已知信息中的图片、链接地址和脚本语言请直接返回。\n- 请使用与问题相同的语言来回答。\n问题：\n{question}\n"
)

// ApplicationCreateReq
type (
	ApplicationCreateReq struct {
		Name                   string         `json:"name"`
		Desc                   string         `json:"desc"`
		ModelID                string         `json:"model_id"`
		MultipleRoundsDialogue bool           `json:"multiple_rounds_dialogue"`
		Prologue               string         `json:"prologue"`
		DatasetIDList          []string       `json:"dataset_id_list"`
		DatasetSetting         DatasetSetting `json:"dataset_setting"`
		ModelSetting           ModelSetting   `json:"model_setting"`
		ProblemOptimization    bool           `json:"problem_optimization"`
	}

	DatasetSetting struct {
		TopN                   int                 `json:"top_n"`
		Similarity             float64             `json:"similarity"`
		MaxParagraphCharNumber int                 `json:"max_paragraph_char_number"`
		SearchMode             string              `json:"search_mode"`
		NoReferencesSetting    NoReferencesSetting `json:"no_references_setting"`
	}

	NoReferencesSetting struct {
		Status string `json:"status"`
		Value  string `json:"value"`
	}

	ModelSetting struct {
		Prompt string `json:"prompt"`
	}
)

// ApplicationCreateRes
type ApplicationCreateRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    bool   `json:"data"`
}

func (app ApplicationCreateReq) DocSetCreateRes2ApplicationCreateReq(req DocSetCreateRes) (res ApplicationCreateReq) {

	res = ApplicationCreateReq{
		Name:                   req.Data.Name,
		Desc:                   req.Data.Desc,
		MultipleRoundsDialogue: true,
		Prologue:               PROLOGUE,
		DatasetIDList:          []string{req.Data.ID},
		ProblemOptimization:    true,
	}

	dataSetSetting := DatasetSetting{
		TopN:                   3,
		Similarity:             0.6,
		MaxParagraphCharNumber: 5000,
		SearchMode:             "embedding",
	}

	noReferencesSetting := NoReferencesSetting{
		Status: "ai_questioning",
		Value:  "{question}",
	}

	dataSetSetting.NoReferencesSetting = noReferencesSetting

	modelSetting := ModelSetting{
		Prompt: PROMPT,
	}

	res.DatasetSetting = dataSetSetting
	res.ModelSetting = modelSetting
	res.ProblemOptimization = false

	return
}
