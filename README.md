- [ ]百度流式输出
- [x]阿里流式输出
- [x]deepseek流式输出
- [x]星火流式输出
### 阿里太坑了，用支持流式的参数，然后发现返回的内容有重复，比如第一句是```这是个例子```，第二句会是```这是个例子，例子的内容```，每一行的输出都包含前面n行的内容,如果做流式还要在内容的重新切割，不然返回的结果有大量重复的内容，放弃了
  
