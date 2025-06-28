" ~/.vim/syntax/ain.vim
if exists("b:current_syntax")
  finish
endif

" Headings
syntax match ainHeading /^\s*\[\(config\|host\|query\|headers\|method\|body\|backend\|backendoptions\)\]\s*\ze\(\s*#\|\s*$\)\c/
highlight link ainHeading Keyword

" Escapes
syntax match ainEscape /`#/
highlight link ainEscape Normal

" Comments
syntax match ainComment /#.*/
highlight link ainComment Comment

let b:current_syntax = "ain"
