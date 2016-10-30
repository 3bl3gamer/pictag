require 'ostruct'
require 'time'
require 'json'
require 'exif' #gem install exif
#exiftool ./mb_dup__101MSDCF_1/DSC06463.JPG | grep Create
#a=[]; for (let i in localStorage) if (i!='length') a.push([i,localStorage[i]]); a.sort((a,b) => String.localeCompare(a,b)); h={}; a.forEach(([k,v]) => h[k]=v); copy(JSON.stringify(h, null, "\t"))

# TODO: для тегов используется комбинация имени и даты, для всех
#       копий одной фотографии будут храниться одни и те же теги,
#       но но обновления страницы этого не будет видно

width = 192
height = 144

cache = JSON.load File.read "cache.json" rescue Errno::ENOENT
cache ||= {}
cache.each{|path,val| val["created_at"] = Time.parse val["created_at"] }

files = Dir.glob("**/*").map do |filepath|
  if [".jpg", ".jpeg"].include?(File.extname(filepath).downcase) and not filepath.start_with? "thumbnails"
    if cache[filepath]
      created_at = cache[filepath]["created_at"]
    else
      begin
        exif = Exif::Data.new(filepath)
        created_at = [exif.date_time_original, exif.date_time_digitized].min
      rescue RuntimeError => e
        raise e unless e.message.include? "no EXIF data in file"
      end
      if created_at and created_at.year == 2079
        created_at = nil #какой-то баг с таймстампаи в exif'е, в File.stat более-менее правильная дата
      end
      unless created_at
        stat = File.stat(filepath)
        created_at = [stat.ctime, stat.mtime].min
      end
      cache[filepath] = {"created_at" => created_at}
    end
    stamp = created_at.strftime "%Y-%m-%d %H:%M:%S"
    name = File.basename(filepath)
    key = "#{stamp} #{name}"
    print "."
    OpenStruct.new(:name => name,
                   :path => filepath,
                   :created_at => created_at,
                   :key => key,
                   :thumb => "thumbnails/#{stamp} #{filepath.gsub("/", " ")}",
                   :dup_group_num => nil)
  end
end.compact.sort_by{|f| f.created_at }.reverse

File.write "cache.json", JSON.pretty_generate(cache, :indent => "\t")

files.group_by{|f| f.key }.select{|time,files| files.size > 1 }.each_with_index{|tf,i| tf[1].each{|f| f.dup_group_num = i } }


files.each do |file|
  `ffmpeg -i "#{file.path}" -filter:v "scale=iw*min(#{width}/iw\\,#{height}/ih):ih*min(#{width}/iw\\,#{height}/ih)" "#{file.thumb}"` unless File.exist?(file.thumb)
end


File.write("index.html",
%Q[
<meta charset="utf-8">
<body>
<style>
  .item {
    position: relative;
    float: left;
    width: 128px;
    height: 128px;
    margin: 3px;
    font-size: 8pt;
    overflow: hidden;
  }
  .item:hover {
    outline: 2px solid gray;
  }
  .item.selected {
    background: gray;
  }
  .item img {
    display: block;
    max-width: 128px;
    max-height: 128px;
    margin: 0 auto;
  }
  #tagsEditorWrap {
    position: fixed;
    left: 8px;
    top: 8px;
    transition: opacity 0.3s ease;
  }
  #tagsEditorWrap:not(:hover):not(.active) {
    opacity: 0.6;
  }
  #autoCompleteBox {
    background-color: white;
    border: 1px solid gray;
  }
  #autoCompleteBox div.selected {
    background-color: lightblue;
  }
  .tags {
    position: absolute;
    width: 100%;
    bottom: 0px;
    background-color: rgba(255, 255, 255, 0.5);
  }
  .tag::after {
    content: " ";
  }
</style>
] + files.each_with_index.map do |file,i|
%Q[
<div class="item" data-i="#{i}" data-path="#{file.path}" data-name="#{file.name}" data-key="#{file.key}">
  <img src="#{file.thumb}" title="#{file.name} #{file.created_at}">
  <div class="tags"></div>
</div>
]
end.join("\n") + %Q[
<div id="tagsEditorWrap">
  <textarea id="tagsArea"></textarea>
  <div id="autoCompleteBox">qwe</div>
  <button id="setTagsButton">s</button>
  <button id="delTagsButton">-</button>
  <button id="addTagsButton">+</button>
</div>
<script>
var allTags = new Set()
var lastSelectedItem = null

document.body.onclick = function(e){
  var itemElem = findItemFrom(e.target)
  if (itemElem) select(e, itemElem)
}
document.body.ondblclick = function(e){
  var itemElem = findItemFrom(e.target)
  window.open(itemElem.dataset.path, '_blank').focus()
}
setTagsButton.onclick = setEnteredTags
delTagsButton.onclick = delEnteredTags
addTagsButton.onclick = addEnteredTags
window.onkeypress = function(e){
  if (!e.shiftKey && e.ctrlKey && e.altKey) {
    if (e.key == "0"){ e.preventDefault(); setEnteredTags() }
    if (e.key == "-"){ e.preventDefault(); delEnteredTags() }
    if (e.key == "="){ e.preventDefault(); addEnteredTags() }
  }
}
tagsArea.onkeydown = function(e) {
  tagsEditorWrap_activate()
  if (e.key == "Tab") {e.preventDefault(); autocompleteTag()}
}
tagsArea.onkeyup = updateAutocomplete

function autocompleteTag() {
  var elem = autoCompleteBox.querySelector('div.selected')
  if (!elem) return
  var beforeSelection = tagsArea.value.substr(0, tagsArea.selectionStart)
  var afterSelection = tagsArea.value.substr(tagsArea.selectionEnd)
  
  var tag = elem.textContent
  var cur = beforeSelection.match(/[^\\n]*$/)[0]
  if (tag == cur) {
    var nextElem = elem.nextElementSibling ? elem.nextElementSibling : autoCompleteBox.children[0]
    elem.classList.remove('selected')
    nextElem.classList.add('selected')
    tag = nextElem.textContent
  }
  
  var before = beforeSelection.match(/^([\\d\\D]*?)[^\\n]*$/)[1]
  var after = afterSelection.match(/^[^\\n]*([\\d\\D]*)$/)[1]
  tagsArea.value = before + tag + after
  tagsArea.selectionStart = tagsArea.selectionEnd = before.length + tag.length
}
function updateAutocomplete(e) {
  if (e.key == "Tab") return
  if (tagsArea.selectionStart != tagsArea.selectionEnd) return
  var str = tagsArea.value.substr(0, tagsArea.selectionStart).match(/[^\\n]*$/)[0]
  var tags = getCompletionTags(str).slice(0, 10)
  autoCompleteBox.innerHTML = tags.map(t => `<div>${t}</div>`).join("")
  if (autoCompleteBox.children.length > 0)
    autoCompleteBox.children[0].classList.add('selected')
}
function getCompletionTags(str) {
  var tags = []
  for (var tag of allTags) if (tag.startsWith(str)) tags.push(tag)
  return tags.reverse()
}

var tagsEditorWrap_active_timeout = -1
function tagsEditorWrap_activate() {
  tagsEditorWrap.classList.add('active')
  clearTimeout(tagsEditorWrap_active_timeout)
  tagsEditorWrap_active_timeout = setTimeout(tagsEditorWrap_deactivate, 2000)
}
function tagsEditorWrap_deactivate() {
  tagsEditorWrap.classList.remove('active')
}

function findItemFrom(elem) {
  while(elem) {
    if (elem.classList.contains('item')) return elem
    elem = elem.parentElement
  }
  return null
}

function addEnteredTags() {
  var itemElems = document.querySelectorAll('.item.selected')
  var tags = getEnteredTags()
  for (var i=0; i<itemElems.length; i++) setTags(itemElems[i], getTags(itemElems[i]).concat(tags))
}
function setEnteredTags() {
  var itemElems = document.querySelectorAll('.item.selected')
  var tags = getEnteredTags()
  for (var i=0; i<itemElems.length; i++) setTags(itemElems[i], tags)
}
function delEnteredTags() {
  var itemElems = document.querySelectorAll('.item.selected')
  var tags = getEnteredTags()
  for (var i=0; i<itemElems.length; i++) setTags(itemElems[i], getTags(itemElems[i]).filter(t => tags.indexOf(t)==-1))
}

function select(e, itemElem) {
  if (e.shiftKey && lastSelectedItem) {
    toggleAllBetween(lastSelectedItem, itemElem, !itemElem.classList.contains('selected'))
  } else {
    if (!e.ctrlKey) unselectAll()
    itemElem.classList.toggle('selected')
    lastSelectedItem = itemElem
  }
  update()
}
function toggleAllBetween(elem1, elem2, on) {
  if (+elem1.dataset.i > +elem2.dataset.i) { var t=elem1; elem1=elem2; elem2=t }
  while (true) {
    elem1.classList.toggle('selected', on)
    if (elem1 == elem2) break
    elem1 = elem1.nextElementSibling
  }
}
function unselectAll() {
  var itemElems = document.querySelectorAll('.item.selected')
  Array.prototype.forEach.call(itemElems, i => i.classList.remove('selected'))
}

function update() {
  var itemElems = document.querySelectorAll('.item.selected')
  switch (itemElems.length) {
  case 0:
    tagsArea.value = ''
    break
  case 1:
    tagsArea.value = getTags(itemElems[0]).join('\\n')
    break
  }
  tagsArea.focus()
  tagsEditorWrap_activate()
  autoCompleteBox.innerHTML = ''
}

function getTags(item) {
  return Array.prototype.map.call(item.querySelectorAll('.tag'), i => i.textContent)
}
function setTags(item, tags) {
  tags = Array.from(new Set(tags)).sort()
  item.querySelector('.tags').innerHTML = tags.map(t => `<span class="tag">${t.trim()}</span>`).join('')
  tags.forEach(t => {allTags.delete(t); allTags.add(t)})
  localStorage[item.dataset.key] = tags.join('\\n')
}
function getEnteredTags() {
  return tagsArea.value.trim().split('\\n').map(t => t.trim()).filter(t => t != '')
}

(function(){
  var itemElems = document.querySelectorAll('.item')
  Array.prototype.forEach.call(itemElems, item => {
    var tags_str = localStorage[item.dataset.key]
    if (tags_str) setTags(item, tags_str.split('\\n'))
  })
})()
</script>
])

