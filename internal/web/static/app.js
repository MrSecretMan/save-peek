const $ = id => document.getElementById(id)
let selected = localStorage.getItem('savepeek-save') || ''

function money(value) { return new Intl.NumberFormat().format(value || 0) + 'g' }
function duration(ms) {
  const hours = Math.floor((ms || 0) / 3600000)
  const minutes = Math.floor(((ms || 0) % 3600000) / 60000)
  return hours + 'h ' + minutes + 'm'
}
function when(value) { return new Date(value).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' }) }

function render(data) {
  $('hero').classList.remove('skeleton')
  $('source').textContent = 'watching ' + data.source.folder
  $('farm').textContent = data.farm_name || data.source.folder
  $('date').textContent = (data.season || '?') + ' ' + (data.day || '?') + ', Year ' + (data.year || '?')
  $('money').textContent = money(data.money)
  $('player').textContent = data.player_name || 'unknown'
  $('achievements').textContent = data.achievements
  $('playtime').textContent = duration(data.play_time_ms)
  $('updated').textContent = when(data.source.modified_at)

  const recs = data.recommendations || []
  $('next').hidden = recs.length === 0
  if (recs.length) {
    $('next-title').textContent = recs[0].title
    $('next-reason').textContent = recs[0].reason
    $('next-more').innerHTML = recs.slice(1).map(rec => '<div class="next-item"><b>' + rec.title + '</b><span>' + rec.reason + '</span></div>').join('')
  }

  const skills = [['farming',data.skills.farming],['mining',data.skills.mining],['fishing',data.skills.fishing],['foraging',data.skills.foraging],['combat',data.skills.combat]]
  $('skills').innerHTML = skills.map(pair => '<div class="skill"><span>' + pair[0] + '</span><span class="track"><i style="width:' + Math.min(100,(pair[1]||0)*10) + '%"></i></span><b>' + (pair[1]||0) + '</b></div>').join('')
  if (data.relationships?.length) {
    $('friends').innerHTML = data.relationships.map(friend => '<div class="friend"><span>' + friend.name + '</span><span>' + friend.hearts + ' ♥</span></div>').join('')
  }
}

async function loadSaves() {
  const response = await fetch('/api/saves', { cache: 'no-cache' })
  if (!response.ok) throw new Error('HTTP ' + response.status)
  const saves = await response.json()
  if (!saves.length) throw new Error('no readable saves')
  if (!saves.some(save => save.id === selected)) selected = saves[0].id
  $('save-select').innerHTML = saves.map(save => {
    const name = save.farm_name || save.player_name || save.id
    return '<option value="' + save.id + '">' + name + ' · ' + (save.season || '?') + ' ' + (save.day || '?') + ', Y' + (save.year || '?') + '</option>'
  }).join('')
  $('save-select').value = selected
}

async function load() {
  $('error').hidden = true
  try {
    if (!$('save-select').options.length) await loadSaves()
    const response = await fetch('/api/save?id=' + encodeURIComponent(selected), { cache: 'no-cache' })
    if (!response.ok) throw new Error('HTTP ' + response.status)
    render(await response.json())
  } catch (err) {
    $('error').textContent = 'Could not read the save: ' + err.message
    $('error').hidden = false
  }
}

$('save-select').addEventListener('change', event => {
  selected = event.target.value
  localStorage.setItem('savepeek-save', selected)
  load()
})
$('refresh').addEventListener('click', load)
load()
