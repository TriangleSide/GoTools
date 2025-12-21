/*
Package heap provides a binary heap implementation.

A binary tree can be represented in an array by having the children be at predefined indexes.

  - Left child: (2*i)+1.
  - Right child: (2*i)+2.
  - Parent: floor((i-1)/2).

Take the following tree:

	       10
	      /  \
	     5    15
	         /  \
	        13   18

	This has the array representation [10, 5, 15, -, -, 13, 18].
	                          Indexes: 0,  1, 2,  3, 4, 5,  6.

A complete binary tree is when nodes in the tree are filled left to right.
This ensures the array representation does not have "gaps".

	    5 - Complete         5 - Incomplete
	   / \                  / \
	  3   8                3   8
	 / \                      / \
	2   4                    7   9

A heap can be represented as a complete binary tree. Using a complete binary tree ensures
the tree is always balanced. The key property of this representation is that each node
must be bigger (if using max heap) than its children. It does not matter if the children
are sorted. This ensures the root node is always the largest.

	  10
	 /  \
	8    5

This tree has the array representation: [10, 8, 5].

When pushing a value (15) to the max heap, it is appended to the end of the array.
The array becomes: [10, 8, 5, 15]. This, however, breaks the property of the heap
where the root is the largest. To fix this, the value is "bubbled" up to the top
of the heap. If the value is larger than its parent, they are swapped, and
the operation is performed again.

1. Append the new element to the end of the array.

2. Bubble up the new value recursively.

	    10      ->     10      ->       15
	   /  \           /  \             /  \
	  8    5         8    5           10   5
	 /                /
	15               8

When popping the top value, it is first swapped with the last element in the array
representation. Then, the array is resized to remove it. This means the array representation
of the above tree goes from [15, 10, 5, 8] to [8, 10, 5]. This, however, breaks the
property of the heap where the root is the largest. To fix this, the node is "bubbled"
down. The largest of the children (if any) is picked to swap.

1. Swap the root with the last element in the array (15 and 8).

2. Remove the last element in the array representation.

3. Bubble down the value recursively.

	    15     ->      8     ->      8       ->     10
	   /  \           / \           / \            /  \
	  10   5         10  5         10  5          8    5
	 /              /
	8              15
*/
package heap
