# crud


### the filter syntax
* Format: `[^]<field list>#<options>`
* Example
  * `""`: default, for all field and skip nil,zero value
  * `"#all"`: for all field and not skip nil,zero value
  * `"#nil"`: for all field and only skip nil value
  * `"#zero"`: for all field and only skip zero value
  * `"tid,name"`: for only include field tid,name and auto skip nil,zero value
  * `"tid,name#all"`: for only include field tid,name and not skip nil,zero value
  * `"^tid,name"`: for exclude field tid,name and auto skip nil,zero value
  * `"^tid,name#all"`: for exclude field tid,name and auto skip nil,zero value
