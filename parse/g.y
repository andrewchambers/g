%{
package parse
%}
%token IDENTIFIER CONSTANT STRING_LITERAL FUNC RETURN


%start translation_unit


%union {
    astNode ASTNode
}


%%


translation_unit : func_def

func_def :
FUNC IDENTIFIER '(' ')' '{' RETURN ';' '}'

%%
