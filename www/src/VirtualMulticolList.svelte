<script>
	import { onMount, tick } from 'svelte'

	export let items
	export let itemWidth
	export let itemHeight

	let visibleItems = []
	let listItemsWrapElem = null
	let numCols = 0
	let numRows = 0
	let firstVisibleRow = 0
	let lastVisibleRow = 0

	function onResize() {
		let rect = listItemsWrapElem.getBoundingClientRect()
		numCols = Math.floor(rect.width / itemWidth)
		numRows = Math.ceil(items.length / numCols)
		onScroll()
	}
	function onScroll() {
		let rect = listItemsWrapElem.getBoundingClientRect()
		let docHeight = document.documentElement.clientHeight
		firstVisibleRow = Math.max(0, Math.floor(-rect.top / itemHeight))
		lastVisibleRow = Math.max(0, Math.ceil((-rect.top + docHeight) / itemHeight))
		let visibleRows = lastVisibleRow - firstVisibleRow
		// firstVisibleRow = Math.max(0, firstVisibleRow - visibleRows)
		// lastVisibleRow += visibleRows
		visibleItems = items.slice(numCols * firstVisibleRow, numCols * lastVisibleRow)
	}

	onMount(onResize)
</script>

<style>
	.list-items-wrap {
		position: relative;
		width: 100%;
		overflow: hidden;
	}
</style>

<svelte:window on:resize={onResize} on:scroll|passive={onScroll} />

<div class="list-items-wrap" bind:this={listItemsWrapElem} style="height:{numRows * itemHeight}px">
	{#each visibleItems as item, i (item.key)}
		<slot
			{item}
			left={(i % numCols) * itemWidth}
			top={(Math.floor(i / numCols) + firstVisibleRow) * itemHeight} />
	{/each}
</div>
