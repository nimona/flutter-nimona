Pod::Spec.new do |s|
  s.name                = 'nimona'
  s.version             = '0.0.6'
  s.summary             = 'Nimona Flutter Plugin.'
  s.description         = 'Nimona Flutter Plugin.'
  s.homepage            = 'https://github.com/nimona/flutter-nimona'
  s.license             = { :file => '../LICENSE' }
  s.author              = { 'George Antoniadis' => 'george@noodles.gr' }
  s.source              = { :path => '.' }
  s.source_files        = 'Classes/**/*'
  s.dependency            'FlutterMacOS'
  s.platform            = :osx, '10.11'
  s.pod_target_xcconfig = { 'DEFINES_MODULE' => 'YES' }
  s.swift_version       = '5.0'
  s.vendored_libraries  = '*.a'
  s.xcconfig            = { 'OTHER_LDFLAGS' => '-force_load "${PODS_ROOT}/../Flutter/ephemeral/.symlinks/plugins/nimona/macos/libnimona.a"'}
end
