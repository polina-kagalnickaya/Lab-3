package dto

type YandexUserInfo struct {
    ID        string `json:"id"`
    Login     string `json:"login"`
    RealName  string `json:"real_name"`
    Email     string `json:"default_email"`
}

type VkUserInfo struct {
    ID        int    `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
}