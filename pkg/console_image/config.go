package console_image

const ResizeOffsetY = 8
const ResizeFactorY = 2
const ResizeFactorX = 1
const DefaultTermCols = 80
const DefaultTermRows = 24
const FPS = 15

const AnsiCursorUp = "\x1B[%dA"
const AnsiCursorHide = "\x1B[?25l"
const AnsiCursorShow = "\x1B[?25h"
const AnsiBgTransparentColor = "\x1b[0;39;49m"
const AnsiBgRgbColor = "\x1b[48;2;%d;%d;%dm"
const AnsiFgTransparentColor = "\x1b[0m "
const AnsiFgRgbColor = "\x1b[38;2;%d;%d;%dm▄"
const AnsiReset = "\x1b[0m"
