package app

type ItemContainer []*Item

func NewItemContainer(size int) ItemContainer {
	slice := make([]*Item, size)
	for i := range slice {
		slice[i] = &Item { -1, 0 }
	}
	return slice
}

func (ic ItemContainer) Add(item *Item) bool {
	for idx, i := range ic {
		if i.ID == -1 {
			ic[idx] = item
			return true
		}
	}
	return false
}

func (ic ItemContainer) RemoveFirst(id int) bool {
	for idx, i := range ic {
		if i.ID == id {
			ic[idx] = &Item{ -1, 0 }
			return true
		}
	}
	return false
}