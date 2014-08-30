%{
package parse
%}

%token ERROR IDENTIFIER CONSTANT STRING_LITERAL FUNC RETURN PACKAGE STRUCT IMPORT FOR


%start translation_unit


%union {
    astNode ASTNode
}


%%


translation_unit : 
    package_decl import_list func_def

package_decl:
    PACKAGE IDENTIFIER

import_list:
    | import import_list
    
import:
    IMPORT STRING_LITERAL

func_def :
    FUNC IDENTIFIER '(' ')' type '{' statement '}'
    FUNC IDENTIFIER '(' ')' '{' statement '}'

type :
    | IDENTIFIER
    | struct

struct :
    STRUCT '{' '}'

statement:
    RETURN expression ';'

expression :
    primary_expression

primary_expression:
    CONSTANT

%%
