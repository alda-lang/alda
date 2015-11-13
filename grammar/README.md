For optimized performance, Alda uses multiple parsers to parse different parts of a score.

The grammars in this directory are not each complete grammars; Instaparse will fail to build most of them individually because they may refer to rules defined in the other grammars. In `src/alda/parser.clj`, parsers are built using different combinations of the grammars here.
