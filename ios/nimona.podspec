Pod::Spec.new do |s|
  s.name                = 'nimona'
  s.version             = '0.0.6'
  s.summary             = 'Nimona Flutter Plugin.'
  s.description         = 'Nimona Flutter Plugin.'
  s.homepage            = 'https://github.com/nimona/flutter-nimona'
  s.license             = { :file => '../LICENSE' }
  s.author              = { 'George Antoniadis' => 'george@noodles.gr' }
  s.source              = { :path => '.' }
  s.source_files        =  'Classes/**/*'
  s.public_header_files = 'Classes/**/*.h'
  s.vendored_libraries  = '*.a'
  s.dependency            'Flutter'
  s.platform            = :ios, '8.0'
  s.xcconfig            = { 'OTHER_LDFLAGS' => '-force_load "${PODS_ROOT}/../.symlinks/plugins/nimona/ios/libnimona.a"'}
  s.pod_target_xcconfig = {  'DEFINES_MODULE' => 'YES', 'VALID_ARCHS[sdk=iphonesimulator*]' => 'x86_64' }
  s.swift_version       = '5.0'
end