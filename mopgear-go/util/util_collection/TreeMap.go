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
type ITreeNode[K Sortable[K], N ITreeNode[K, N]] interface {
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
type treeBase[K Sortable[K], N ITreeNode[K, N]] struct {
	root N
}

//func (tb *treeBase[K, N]) findNode(key K) (N, bool) {
//	curr := tb.root
//	for curr != nil {
//		x := key.CompareTo(curr.GetKey())
//		if x < 0 {
//			curr = curr.GetLeft()
//		} else if x > 0 {
//			curr = curr.GetRight()
//		} else {
//			return curr, true
//		}
//	}
//	return makeNilValue[N](), false
//}
//
//func (tb *treeBase[K, N]) insertSearch(add N) {
//	key := add.GetKey()
//	curr := tb.root
//
//	if curr == nil {
//		tb.root = add
//		return
//	}
//
//	for curr != nil {
//		x := key.CompareTo(curr.GetKey())
//		if x < 0 {
//			if curr.GetLeft() == nil {
//				curr.SetLeft(add)
//				tb.finishInsert(add, curr)
//			} else {
//				curr = curr.GetLeft()
//			}
//		} else if x > 0 {
//			if curr.GetRight() == nil {
//				curr.SetRight(add)
//				tb.finishInsert(add, curr)
//			} else {
//				curr = curr.GetRight()
//			}
//		} else {
//			panic("duplicate")
//		}
//	}
//}
//
//func (tb *treeBase[K, N]) directionIsRight(node N) bool {
//	return node == node.GetParent().GetRight()
//}
//
//func (tb *treeBase[K, N]) nodeChildByDirection(node N, directionIsRight bool) N {
//	if directionIsRight {
//		return node.GetRight()
//	} else {
//		return node.GetLeft()
//	}
//}
//
////https://en.wikipedia.org/wiki/Red%E2%80%93black_tree
//func (tb *treeBase[K, N]) finishInsert(add N, parent N) {
//	add.SetParent(parent)
//
//	for {
//		if parent.IsBlack() {
//			return
//		}
//
//		grandParent := parent.GetParent()
//		if grandParent == nil {
//			parent.SetBlack(true)
//			return
//		}
//
//		parentIsRight := tb.directionIsRight(parent)
//		uncle := tb.nodeChildByDirection(grandParent, !parentIsRight)
//		if uncle == nil || uncle.IsBlack() {
//			if add == tb.nodeChildByDirection(parent, !parentIsRight) {
//
//			}
//
//		}
//	}
//}

// ----------------------------------------- //
type TreeMap[K Sortable[K], V any] struct {
	treeBase[K, *treeMapNode[K, V]]
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
