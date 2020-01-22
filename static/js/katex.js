
function renderKatex(element) {
  renderMathInElement(element, {
    delimiters: [
      { left: "$", right: "$", display: false },
      { left: "$$", right: "$$", display: true },
      { left: "\\(", right: "\\)", display: false },
      { left: "\\[", right: "\\]", display: true }
    ],
    throwOnError: false,
  })
}

function initKatex() {
  $('.commont-content, .topic-content, #id_preview').each(function (e) {
    renderKatex(e)
  })
}

function renderKatexInPreview() {
  $('#id_preview').each(function (e) {
    renderKatex(e)
  })
}
