" ~/.vim/syntax/ain.vim
if exists("b:current_syntax")
  finish
endif

" Headings
syntax match ainHeading /^\s*\[\(config\|host\|query\|headers\|method\|body\|backend\|backendoptions\)\]\s*\ze\(\s*#\|\s*$\)\c/
highlight link ainHeading Keyword

" Escapes
syntax match ainEscape /\\`/
syntax match ainEscape /`\${/
syntax match ainEscape /`#/
highlight link ainEscape Normal

" Envvars: ${VAR}
syntax region ainEnvvar start=+\${+ end=+}+ contains=ainEnvvarEscape
syntax match ainEnvvarEscape /`}/ contained
highlight link ainEnvvar Keyword
highlight link ainEnvvarEscape Keyword

" Comments
syntax match ainComment /#.*/
highlight link ainComment Comment

let b:current_syntax = "ain"
