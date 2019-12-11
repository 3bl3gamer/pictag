<script>
	export let allTags = []
	export let onSetTags
	export let onAddTags
	export let onDelTags
	export function focusOn(newTags) {
		tags = newTags
		if (tagsArea) {
			tagsArea.value = tags.join('\n')
			tagsArea.focus()
			tagsArea.scrollTop = tagsArea.scrollHeight
			activateEditor()
		}
	}
	export function blur() {
		tagsArea.value = ''
		deactivateEditor()
	}

	let tags = []
	let isActive = false
	let tagsEditorWrap_activeTimeout = null
	let completionTags = []
	let selectedCompTagNum = null
	let tagsArea = null

	function activateEditor() {
		isActive = true
		clearTimeout(tagsEditorWrap_activeTimeout)
	}
	function deactivateEditor() {
		isActive = false
		clearTimeout(tagsEditorWrap_activeTimeout)
	}
	function deactivateEditorDelayed() {
		clearTimeout(tagsEditorWrap_activeTimeout)
		tagsEditorWrap_activeTimeout = setTimeout(deactivateEditor, 1000)
	}

	function autocompleteTag() {
		let beforeSelection = tagsArea.value.substr(0, tagsArea.selectionStart)
		let afterSelection = tagsArea.value.substr(tagsArea.selectionEnd)

		let tag = completionTags[selectedCompTagNum]
		let cur = beforeSelection.match(/[^\n]*$/)[0]
		if (tag == cur) {
			selectedCompTagNum = (selectedCompTagNum + 1) % completionTags.length
			tag = completionTags[selectedCompTagNum]
		}

		let before = beforeSelection.match(/^([\d\D]*?)[^\n]*$/)[1]
		let after = afterSelection.match(/^[^\n]*([\d\D]*)$/)[1]
		tagsArea.value = before + tag + after
		tagsArea.selectionStart = tagsArea.selectionEnd = before.length + tag.length
	}

	function updateAutocomplete() {
		if (tagsArea.selectionStart != tagsArea.selectionEnd) return
		let str = tagsArea.value.substr(0, tagsArea.selectionStart).match(/[^\n]*$/)[0]
		completionTags = getCompletionTags(str)
		selectedCompTagNum = completionTags.length > 0 ? 0 : null
	}
	function getCompletionTags(str) {
		let tags = []
		for (let tag of allTags) if (tag.startsWith(str)) tags.push(tag)
		return tags.slice(-10).reverse()
	}

	function getCurrentTags() {
		return tagsArea.value
			.split('\n')
			.map(x => x.trim())
			.filter(x => x != '')
	}

	function onAreaKeyDown(e) {
		activateEditor()
		if (e.key == 'Tab') {
			e.preventDefault()
			autocompleteTag()
		} else if (e.key == 'Escape') {
			tagsArea.blur()
			deactivateEditor()
		}
	}
	function onAreaKeyUp(e) {
		if (e.key != 'Tab') updateAutocomplete()
	}

	function onKeyUp(e) {
		if (!e.shiftKey && e.ctrlKey && e.altKey) {
			let map = { '0': onSetTags, '-': onDelTags, '=': onAddTags }
			if (e.key in map) {
				map[e.key](getCurrentTags())
				e.preventDefault()
			}
		}
	}
	function onTagsButtonClick() {
		let map = { set: onSetTags, del: onDelTags, add: onAddTags }
		map[this.dataset.action](getCurrentTags())
	}
</script>

<style>
	.tagsEditorWrap {
		position: fixed;
		left: 8px;
		top: 8px;
		transition: opacity 0.2s ease;
	}
	.tagsArea {
		height: 38px;
		transition: height 0.2s ease;
	}
	.tagsEditorWrap.active .tagsArea {
		height: 96px;
	}
	.tagsEditorWrap:not(:hover):not(.active) {
		opacity: 0.6;
	}
	.autoCompleteBox {
		background-color: white;
		border: 1px solid gray;
	}
	.autoCompleteBox div.selected {
		background-color: lightblue;
	}
</style>

<svelte:window on:keyup={onKeyUp} />

<div class="tagsEditorWrap {isActive ? 'active' : ''}" on:mouseleave={deactivateEditor}>
	<textarea
		class="tagsArea"
		on:keydown={onAreaKeyDown}
		on:keyup={onAreaKeyUp}
		on:mouseenter={activateEditor}
		on:blur={deactivateEditorDelayed}
		bind:this={tagsArea} />
	<div class="autoCompleteBox">
		{#each completionTags as tag, i (tag)}
			<div class:selected={selectedCompTagNum == i}>{tag}</div>
		{/each}
	</div>
	<button class="setTagsButton" data-action="set" on:click={onTagsButtonClick}>s</button>
	<button class="delTagsButton" data-action="del" on:click={onTagsButtonClick}>-</button>
	<button class="addTagsButton" data-action="add" on:click={onTagsButtonClick}>+</button>
</div>
