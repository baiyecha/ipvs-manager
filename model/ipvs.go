package model

type Ipvs struct{
	VIP string `json:"vip"`
	Backends []*Backend `json:"backends"`
	Protocol string `json:"protocol"`
	SchedName string `json:"sched_name"`
}

type Backend struct{
	Addr string `json:"addr"`
	Weight int `json:"weight"`
	Status int `json:"status"` // ipvs后端的健康状态，1为不健康，0为健康
}

type IpvsList struct{
	IpvsList []*Ipvs `json:"ipvs_list"`
}
