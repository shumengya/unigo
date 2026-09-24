; UniGo highlights (Zed style: later patterns take precedence over earlier ones)

(identifier) @variable

; Types
(type_identifier) @type
(qualified_type
  package: (identifier) @namespace
  name: (identifier) @type)

((type_identifier) @type.builtin
  (#match? @type.builtin "^(bool|byte|rune|string|error|any|int|int8|int16|int32|int64|uint|uint8|uint16|uint32|uint64|uintptr|float32|float64|complex64|complex128)$"))

; Declarations
(package_clause name: (identifier) @namespace)
(import_declaration name: (identifier) @namespace)

(enum_member (identifier) @constant)
(const_declaration name: (identifier) @constant)

(field_declaration name: (identifier) @property)
(field_initializer name: (identifier) @property)
(selector_expression field: (identifier) @property)

(parameter name: (identifier) @variable.parameter)

(function_declaration name: (identifier) @function)
(method_signature name: (identifier) @function.method)
(method_implementation name: (identifier) @function.method)

(composite_literal type: (identifier) @type)
(composite_literal type: (qualified_type name: (identifier) @type))

; Calls
(call_expression function: (identifier) @function.call)
(call_expression
  function: (selector_expression field: (identifier) @function.method.call))

((call_expression function: (identifier) @function.builtin)
  (#match? @function.builtin "^(delete|len|cap|append|copy|panic|print|println|make|min|max|clear)$"))

; Literals
(int_literal) @number
(float_literal) @number
[(true) (false)] @boolean
(nil) @constant.builtin

(string) @string
(raw_string) @string
(escape_sequence) @string.escape
(import_declaration path: (string) @string.special)

(comment) @comment

; Keywords
[
  "package"
  "import"
  "from"
  "var"
  "const"
] @keyword

"func" @keyword.function

"return" @keyword.return

[
  "if"
  "elif"
  "else"
  "switch"
  "case"
  "default"
] @keyword.conditional

[
  "while"
  "for"
] @keyword.repeat

[
  "enum"
  "struct"
  "interface"
  "newtype"
  "alias"
  "map"
] @keyword.type

; Operators & punctuation
[
  "+"
  "-"
  "*"
  "/"
  "%"
  "=="
  "!="
  "<"
  ">"
  "<="
  ">="
  "&&"
  "||"
  "!"
  "&"
  "="
] @operator

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
] @punctuation.bracket

[
  ","
  ";"
  ":"
  "."
] @punctuation.delimiter
