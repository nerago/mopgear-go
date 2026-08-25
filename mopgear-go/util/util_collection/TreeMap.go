package util_collection

import "cmp"

func makeNilValue[T any]() T {
	var nilValue T
	return nilValue
}

type OrderedSortable[O cmp.Ordered] struct {
	Value O
}

func (o OrderedSortable[O]) CompareTo(x OrderedSortable[O]) int {
	return cmp.Compare(o.Value, x.Value)
}

var _ Sortable[OrderedSortable[int]] = OrderedSortable[int]{}

// ----------------------------------------- //

type Sortable[K Sortable[K]] interface {
	// CompareTo returns:
	//	-1 if receiver is less than param,
	//	 0 if receiver equals param,
	//	+1 if receiver is greater than param.
	CompareTo(K) int
}

// ----------------------------------------- //
type ITreeNode[K Sortable[K], N ITreeNode[K, N, P], P any] interface {
	*P
	comparable
	GetKey() K
	GetParent() N
	GetLeft() N
	GetRight() N
	SetParent(N)
	SetLeft(N)
	SetRight(N)
	IsBlack() bool
	SetBlack(black bool)
}

// ----------------------------------------- //
type treeBase[K Sortable[K], N ITreeNode[K, N, P], P any] struct {
	root N
}

func (tb *treeBase[K, N, P]) findNode(key K) (N, bool) {
	curr := tb.root
	for curr != nil {
		x := key.CompareTo(curr.GetKey())
		if x < 0 {
			curr = curr.GetLeft()
		} else if x > 0 {
			curr = curr.GetRight()
		} else {
			return curr, true
		}
	}
	return makeNilValue[N](), false
}

func (tb *treeBase[K, N, P]) insertSearch(add N) {
	key := add.GetKey()
	curr := tb.root

	if curr == nil {
		tb.root = add
		return
	}

	for curr != nil {
		x := key.CompareTo(curr.GetKey())
		if x < 0 {
			if curr.GetLeft() == nil {
				curr.SetLeft(add)
				tb.finishInsert(add, curr)
			} else {
				curr = curr.GetLeft()
			}
		} else if x > 0 {
			if curr.GetRight() == nil {
				curr.SetRight(add)
				tb.finishInsert(add, curr)
			} else {
				curr = curr.GetRight()
			}
		} else {
			panic("duplicate")
		}
	}
}

func isRightChild[N ITreeNode[K, N, P], K Sortable[K], P any](node N) bool {
	return node == node.GetParent().GetRight()
}

func getNodeChild[N ITreeNode[K, N, P], K Sortable[K], P any](node N, directionIsRight bool) N {
	if directionIsRight {
		return node.GetRight()
	} else {
		return node.GetLeft()
	}
}

func setNodeChild[N ITreeNode[K, N, P], K Sortable[K], P any](node N, directionIsRight bool, child N) {
	if directionIsRight {
		node.SetRight(child)
	} else {
		node.SetLeft(child)
	}
}

// https://en.wikipedia.org/wiki/Red%E2%80%93black_tree
func (tb *treeBase[K, N, P]) finishInsert(node N, parent N) {
	node.SetParent(parent)

	for {
		if parent.IsBlack() {
			return
		}

		grandParent := parent.GetParent()
		if grandParent == nil {
			parent.SetBlack(true)
			return
		}

		dir := isRightChild[N, K, P](parent)
		uncle := getNodeChild[N, K, P](grandParent, !dir)
		if uncle == nil || uncle.IsBlack() {
			if node == getNodeChild[N, K, P](parent, !dir) {
				tb.rotateSubtree(parent, dir)
				node = parent
				parent = getNodeChild[N, K, P](grandParent, dir)
			}

			tb.rotateSubtree(grandParent, !dir)
			parent.SetBlack(true)
			grandParent.SetBlack(false)
			return
		}

		parent.SetBlack(true)
		uncle.SetBlack(true)
		grandParent.SetBlack(false)
		node = grandParent
		parent = node.GetParent()
		if parent == nil {
			return
		}
	}
}

func (tb *treeBase[K, N, P]) finishInsert2(node N, parent N) {
	node.SetParent(parent)

	for {
		if parent.IsBlack() {
			return
		}

		grandParent := parent.GetParent()
		if grandParent == nil {
			parent.SetBlack(true)
			return
		}

		var uncle N
		if parent == parent.GetParent().GetRight() {
			uncle = grandParent.GetLeft()
			if uncle == nil || uncle.IsBlack() {
				if node == parent.GetLeft() {
					tb.rotateSubtree(parent, true)
					node = parent
					parent = grandParent.GetRight()
				}
				tb.rotateSubtree(grandParent, false)
				parent.SetBlack(true)
				grandParent.SetBlack(false)
				return
			}
		} else {
			uncle = grandParent.GetRight()
			if uncle == nil || uncle.IsBlack() {
				if node == parent.GetRight() {
					tb.rotateSubtree(parent, false)
					node = parent
					parent = grandParent.GetLeft()
				}
				tb.rotateSubtree(grandParent, true)
				parent.SetBlack(true)
				grandParent.SetBlack(false)
				return
			}
		}

		parent.SetBlack(true)
		uncle.SetBlack(true)
		grandParent.SetBlack(false)
		node = grandParent
		parent = node.GetParent()
		if parent == nil {
			return
		}
	}
}

func (tb *treeBase[K, N, P]) rotateSubtree(sub N, dir bool) N {
	subParent := sub.GetParent()
	newRoot := getNodeChild[N, K, P](sub, !dir)
	newChild := getNodeChild[N, K, P](newRoot, dir)

	setNodeChild[N, K, P](sub, !dir, newChild)

	if newChild != nil {
		newChild.SetParent(sub)
	}

	setNodeChild[N, K, P](newRoot, dir, sub)

	newRoot.SetParent(subParent)
	sub.SetParent(newRoot)
	if subParent != nil {
		subIsRight := isRightChild[N, K, P](sub)
		setNodeChild[N, K, P](subParent, subIsRight, newRoot)
	} else {
		tb.root = newRoot
	}

	return newRoot
}

// ----------------------------------------- //
type TreeMap[K Sortable[K], V any] struct {
	treeBase[K, *treeMapNode[K, V], treeMapNode[K, V]]
}

type treeMapNode[K Sortable[K], V any] struct {
	key    K
	value  V
	black  bool
	left   *treeMapNode[K, V]
	right  *treeMapNode[K, V]
	parent *treeMapNode[K, V]
}

//var _ ITreeNode[OrderedSortable[int], *treeMapNode[OrderedSortable[int], string]] = &treeMapNode[OrderedSortable[int], string]{}

func (n *treeMapNode[K, V]) GetKey() K {
	return n.key
}

func (n *treeMapNode[K, V]) GetParent() *treeMapNode[K, V] {
	return n.parent
}

func (n *treeMapNode[K, V]) GetLeft() *treeMapNode[K, V] {
	return n.left
}

func (n *treeMapNode[K, V]) GetRight() *treeMapNode[K, V] {
	return n.right
}

func (n *treeMapNode[K, V]) SetParent(parent *treeMapNode[K, V]) {
	n.parent = parent
}

func (n *treeMapNode[K, V]) SetLeft(left *treeMapNode[K, V]) {
	n.left = left
}

func (n *treeMapNode[K, V]) SetRight(right *treeMapNode[K, V]) {
	n.right = right
}

func (n *treeMapNode[K, V]) IsBlack() bool {
	return n.black
}

func (n *treeMapNode[K, V]) SetBlack(black bool) {
	n.black = black
}

// ----------------------------------------- //
type treeSetNode[K Sortable[K]] struct {
	key    K
	black  bool
	left   *treeSetNode[K]
	right  *treeSetNode[K]
	parent *treeSetNode[K]
}

//var _ ITreeNode[OrderedSortable[int], *treeSetNode[OrderedSortable[int]]] = &treeSetNode[OrderedSortable[int]]{}

func (n *treeSetNode[K]) GetKey() K {
	return n.key
}

func (n *treeSetNode[K]) GetParent() *treeSetNode[K] {
	return n.parent
}

func (n *treeSetNode[K]) GetLeft() *treeSetNode[K] {
	return n.left
}

func (n *treeSetNode[K]) GetRight() *treeSetNode[K] {
	return n.right
}

func (n *treeSetNode[K]) SetParent(parent *treeSetNode[K]) {
	n.parent = parent
}

func (n *treeSetNode[K]) SetLeft(left *treeSetNode[K]) {
	n.left = left
}

func (n *treeSetNode[K]) SetRight(right *treeSetNode[K]) {
	n.right = right
}

func (n *treeSetNode[K]) IsBlack() bool {
	return n.black
}

func (n *treeSetNode[K]) SetBlack(black bool) {
	n.black = black
}
