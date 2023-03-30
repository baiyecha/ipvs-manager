package model

type Ipvs struct{
	VIP string `json:"vip"`
	Backends []Backend `json:"backends"`
	Protocol string `json:"protocol"`
	SchedName string `json:"sched_name"`
}

type Backend struct{
	Addr string `json:"addr"`
	Weight int `json:"weight"`
}

type IpvsList struct{
	IpvsList []*Ipvs `json:"ipvs_list"`
}
