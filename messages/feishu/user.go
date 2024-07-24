package feishu

type User struct {
	Code int `json:"code"`
	Data struct {
		HasMore bool `json:"has_more"`
		Items   []struct {
			Email    string `json:"email"`
			JobTitle string `json:"job_title"`
			Mobile   string `json:"mobile"`
			Name     string `json:"name"`
			UserId   string `json:"user_id"`
		} `json:"items"`
		PageToken string `json:"page_token"`
	} `json:"data"`
	Msg string `json:"msg"`
}
