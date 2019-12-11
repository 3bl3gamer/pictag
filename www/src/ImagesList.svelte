<script>
	import { onMount } from 'svelte'
	import TaggedImageThumb from './TaggedImageThumb.svelte'
	import TagsInput from './TagsInput.svelte'
	import VirtualMulticolList from './VirtualMulticolList.svelte'

	let itemMargin = 3
	let itemWidth = 128 + itemMargin * 2
	let itemHeight = 128 + itemMargin * 2

	let images = []
	let imagesTags = {}
	let allTags = new Set()
	let lastSelectedImage = null
	let selectedImages = new Set()
	let tagsInput = null

	$: isSelected = image => selectedImages.has(image)
	function imageToggle(image, on = null) {
		if (on === null) on = !selectedImages.has(image)
		on ? selectedImages.add(image) : selectedImages.delete(image)
		selectedImages = selectedImages
	}
	function imageUnselectAll() {
		selectedImages.clear()
		selectedImages = selectedImages
	}

	function getImageFromElem(elem) {
		let imageElem = elem.closest('.tagged-image')
		if (imageElem === null) return null
		return images.find(i => i.key == imageElem.dataset.key) || null
	}
	function select(image, isRangeMode, isToggleMode) {
		// imagesTags[image.key] = ['test']
		if (isRangeMode && lastSelectedImage) {
			toggleAllBetween(lastSelectedImage, image, !isSelected(image))
		} else {
			if (!isToggleMode) imageUnselectAll()
			imageToggle(image)
		}
		lastSelectedImage = image
		if (selectedImages.size > 0) {
			tagsInput.focusOn((imagesTags[Array.from(selectedImages)[0].key] || []).slice())
		} else {
			tagsInput.blur()
		}
	}
	function toggleAllBetween(imageFrom, imageTo, on) {
		let indexFrom = images.indexOf(imageFrom)
		let indexTo = images.indexOf(imageTo)
		if (indexFrom > indexTo) [indexFrom, indexTo] = [indexTo, indexFrom]
		for (let i = indexFrom; i <= indexTo; i++) imageToggle(images[i], on)
	}

	$: getImagesTags = image => imagesTags[image.key] || []
	function setImageTags(image, tags) {
		tags = Array.from(new Set(tags)).sort()
		imagesTags[image.key] = tags
		for (let tag of tags) {
			allTags.delete(tag)
			allTags.add(tag)
		}
	}

	onMount(() => {
		fetch('/images')
			.then(r => r.json())
			.then(res => {
				window.res = res.result
				images = res.result.images.slice(7500)
				imagesTags = res.result.tags
				for (let key in imagesTags) {
					let tag = imagesTags[key]
					if (tag != '') {
						let tags = tag.split('\n')
						imagesTags[key] = tags
						for (let i = 0; i < tags.length; i++) allTags.add(tags[i])
					}
				}
			})
	})

	function onItemDblClick(item) {
		window.open('/img/' + item.relativePath, '_blank').focus()
	}
	function onListClick(e) {
		let image = getImageFromElem(e.target)
		if (image !== null) select(image, e.shiftKey, e.ctrlKey)
	}

	function onSetTags(tags) {
		for (let image of selectedImages) setImageTags(image, tags)
	}
	function onDelTags(tags) {
		for (let image of selectedImages)
			setImageTags(image, getImagesTags(image).filter(t => !tags.includes(t)))
	}
	function onAddTags(tags) {
		for (let image of selectedImages) setImageTags(image, getImagesTags(image).concat(tags))
	}
</script>

<style>

</style>

<div class="images-list-wrap" on:click={onListClick}>
	{#if images.length == 0}
		<div>loading...</div>
	{:else}
		<VirtualMulticolList items={images} {itemWidth} {itemHeight} let:item={image} let:left let:top>
			<TaggedImageThumb
				imageItem={image}
				{left}
				{top}
				tags={getImagesTags(image)}
				isSelected={isSelected(image)}
				onOpen={onItemDblClick} />
		</VirtualMulticolList>
	{/if}
	<TagsInput bind:this={tagsInput} {allTags} {onSetTags} {onDelTags} {onAddTags} />
</div>
