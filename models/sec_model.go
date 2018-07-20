package models

type SecModel struct {

}

func (this *SecModel) AllProducts() (prdId int, err error) {
	err   = nil
	prdId = 100
	return
}
