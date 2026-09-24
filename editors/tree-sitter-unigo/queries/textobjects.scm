(function_declaration
  body: (block
    "{"
    (_)* @function.inside
    "}")) @function.around

(method_implementation
  body: (block
    "{"
    (_)* @function.inside
    "}")) @function.around

(method_signature) @function.around

(struct_declaration
  body: (field_list
    "{"
    (_)* @class.inside
    "}")) @class.around

(interface_declaration
  body: (method_list
    "{"
    (_)* @class.inside
    "}")) @class.around

(enum_declaration
  body: (enum_body
    "{"
    (_)* @class.inside
    "}")) @class.around

(comment)+ @comment.around
