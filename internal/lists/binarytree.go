package lists

type BinaryNode struct {
	left  *BinaryNode
	right *BinaryNode
	data  int64
}

type BinaryTree struct {
	root *BinaryNode
}

func (t *BinaryTree) Insert(data int64) *BinaryTree {
	if t.root == nil {
		t.root = &BinaryNode{data: data, left: nil, right: nil}
	} else {
		t.root.insert(data)
	}
	return t
}
func (n *BinaryNode) insert(data int64) {
	if n == nil {
		return
	} else if data <= n.data {
		if n.left == nil {
			n.left = &BinaryNode{data: data, left: nil, right: nil}
		} else {
			n.left.insert(data)
		}
	} else {
		if n.right == nil {
			n.right = &BinaryNode{data: data, left: nil, right: nil}
		} else {
			n.right.insert(data)
		}
	}
}
func (n *BinaryNode) search(data int64) bool {
	//fmt.Printf("verify: %d, %d\n", data, n.data)
	if n.data == data {
		return true
	}
	if n.left != nil && data <= n.data {
		return n.left.search(data)
	}
	if n.right != nil && data >= n.data {
		return n.right.search(data)
	}
	return false
}
func (t *BinaryTree) Search(data int64) bool {
	if t.root == nil {
		return false
	}
	return t.root.search(data)
}
