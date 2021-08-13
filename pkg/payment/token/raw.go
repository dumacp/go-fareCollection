package token

func (p *token) SetRawDataBefore(data interface{}) {
	p.rawDataBefore = data
}

func (p *token) SetRawDataAfter(data interface{}) {
	p.rawDataAfter = data
}

func (p *token) RawDataBefore() interface{} {
	return p.rawDataBefore
}

func (p *token) RawDataAfter() interface{} {
	return p.rawDataAfter
}
