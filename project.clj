(defproject yggdrasil "0.1.0"
  :description "A music programming language for musicians"
  :url "http://www.none-yet.com"  ; will make a github when ready to go public
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :dependencies [[org.clojure/clojure "1.5.1"]
                 [org.clojure/tools.cli "0.2.4"]
                 [instaparse "1.3.3"]]
  :main yggdrasil.core)