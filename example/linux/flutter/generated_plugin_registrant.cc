//
//  Generated file. Do not edit.
//

#include "generated_plugin_registrant.h"

#include <nimona/nimona_plugin.h>

void fl_register_plugins(FlPluginRegistry* registry) {
  g_autoptr(FlPluginRegistrar) nimona_registrar =
      fl_plugin_registry_get_registrar_for_plugin(registry, "NimonaPlugin");
  nimona_plugin_register_with_registrar(nimona_registrar);
}
