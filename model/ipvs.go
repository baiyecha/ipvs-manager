package model

type Ipvs struct{
	VIP string `json:"vip"`
	Backends []string `json:"backends"`
	Rule string `json:"rule"`
}

type IpvsList struct{
	IpvsList []Ipvs `json:"ipvs_list"`
}
