package mplus

func (p *mplus) SetRawDataBefore(data interface{}) {
	p.rawDataBefore = data
}

func (p *mplus) SetRawDataAfter(data interface{}) {
	p.rawDataAfter = data
}

func (p *mplus) RawDataBefore() interface{} {
	return p.rawDataBefore
}

func (p *mplus) RawDataAfter() interface{} {
	return p.rawDataAfter
}
