设计一个验证码校验中转服务，主要验证流程以 cloudflare Turnstile Explicit rendering 方式的校验流程为准，文档地址为：
https://developers.cloudflare.com/turnstile/get-started/server-side-validation/

该服务作用在Process 3、4、5环节，前端会将用户验证完成的token和验证服务提供方（目前为cloudflare）增加到请求header中
由当前服务根据信息对token的有效性进行校验。

设计需要满足以下功能：
1. 可拓展性：后续支持多种校验提供方
2. 提供校验工具：直接提供 middleware 工具包，方便其他服务引入嵌入到接口中完成无侵入接入验证码能力。校验工具对token进行校验时，可选直接接入cloudflare或者接入当前中转服务。
3. 支持多服务中转：允许让校验工具直接将token校验转发至当前服务，由当前服务进行不同的路由校验。
4. 代码的组织方式遵循 golang-standards / project-layout 规范： https://github.com/golang-standards/project-layout, 根据具体需要保留核心内容，无需创建当前阶段不需要不涉及的文件结构

技术栈选择：
1. 中转服务和middleware的交互使用grpc进行，但是健康监测接口不需要通过grpc进行暴露而是继续使用http

其他要求：
1. proto代码需要确保通过protoc命令和构建脚本生成，不要自行生成避免和工具生成结果不一致

进行代码实现和依赖时，中转服务应该尽可能服用校验工具中的直接对接提供商的实现代码，避免重复开发和额外的维护成本

进行技术方案设计和接口文档的设计，将结果输入到reademe文件中