<script>
	export let imageItem
	export let tags
	export let left
	export let top
	export let isSelected
	export let onOpen

	const classTags = ['del', '-90°', '90°', '180°']

	let isLoadng = true
	function onImgLoad() {
		isLoadng = false
	}

	$: tagClasses = classTags
		.filter(t => tags.includes(t))
		.map(t => 'tag-' + t)
		.join(' ')

	function onItemDblClick() {
		onOpen(imageItem)
	}
</script>

<style>
	.tagged-image {
		position: absolute;
		float: left;
		width: 128px;
		height: 128px;
		margin: 3px;
		font-size: 8pt;
		overflow: hidden;
	}
	.tagged-image.loading {
		background-color: lightgray;
	}
	.tagged-image.selected {
		background: gray;
		outline: 3px solid gray;
	}
	.tagged-image:hover {
		outline: 2px solid gray;
	}
	.tagged-image.tag-del:not(:hover) {
		opacity: 0.5;
	}
	.tagged-image.tag--90° img {
		transform-origin: 56px 56px;
		transform: rotate(90deg);
	}
	.tagged-image.tag-90° img {
		transform-origin: 72px 56px;
		transform: rotate(-90deg);
	}
	.tagged-image.tag-180° img {
		transform: rotate(180deg);
	}
	.tagged-image img {
		display: block;
		max-width: 128px;
		max-height: 128px;
		margin: 0 auto;
	}

	.tags {
		position: absolute;
		width: 100%;
		bottom: 0px;
		background-color: rgba(255, 255, 255, 0.5);
	}
	.tagged-image:hover .tag {
		background-color: rgba(0, 0, 0, 0.2);
		border-radius: 4px;
		box-shadow: inset 0px 0px 1px 1px rgba(0, 0, 0, 0.9);
	}
</style>

<div
	class="tagged-image {tagClasses}"
	class:loading={isLoadng}
	class:selected={isSelected}
	style="left:{left}px; top:{top}px"
	data-key={imageItem.key}
	on:dblclick={onItemDblClick}>
	<img
		src="/thumb/{imageItem.relativeThumb}"
		title="{imageItem.name}{' '}{imageItem.stamp}"
		alt={imageItem.name}
		on:load={onImgLoad} />
	<div class="tags">
		{#each tags as tag}
			<span class="tag">{tag}</span>
			{''}
		{/each}
	</div>
</div>
