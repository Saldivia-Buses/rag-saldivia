<?php
/* 
 + Histrix Plugin Interface
 * 2010-01-05
 */

/**
 *
 * @author Luis Melgratti
 */
interface PluginInterface {
    // Initialize Plugin
    public function init();

    // Returns an Array of functions calls to named Hooks
    public function getFunctionHooks();

}
?>
