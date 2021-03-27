#import "NimonaPlugin.h"
#if __has_include(<nimona/nimona-Swift.h>)
#import <nimona/nimona-Swift.h>
#else
// Support project import fallback if the generated compatibility header
// is not copied when this plugin is created as a library.
// https://forums.swift.org/t/swift-static-libraries-dont-copy-generated-objective-c-header/19816
#import "nimona-Swift.h"
#endif

@implementation NimonaPlugin
+ (void)registerWithRegistrar:(NSObject<FlutterPluginRegistrar>*)registrar {
  [SwiftNimonaPlugin registerWithRegistrar:registrar];
}
@end
