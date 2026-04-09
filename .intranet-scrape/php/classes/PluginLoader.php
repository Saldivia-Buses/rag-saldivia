<?php
/**
 * PluginLoader
 * 2010-05-01
 *
 * @author Luis Melgratti
 */
class PluginLoader
{
    public function __construct($pluginPath='../plugins/')
    {
        $this->pluginPath = dirname(__FILE__).'/'.$pluginPath;
    }

    /**
     *
     * @return Array PluginList
     *
     */
    public function registerPlugins()
    {
        return $this->pluginList;
    }

    public function getAvailablePlugins()
    {
        $files = array_diff(scandir($this->pluginPath),array('.','..'));

        foreach ($files as $dirName) {
            if (is_dir($this->pluginPath.$dirName)) {

                $className = $this->pluginPath.$dirName.'/'.$dirName.'.php';
                if (file_exists($className)) {

                    include_once($className);
                    $plugin = new $dirName();

                    $enable = $plugin->loadParameters($this->pluginPath.$dirName.'/');

                    //create default options for plugins
                    if ($enable == -1) {
                        $plugin->createOptions($plugin);
                    }

                    if ($enable == 'true') {

                        $plugin->init();
                        $functionHooks = $plugin->getFunctionHooks();

                        if (is_array($functionHooks))
                        foreach ($functionHooks as $hook => $hooks) {
                            foreach ($hooks as $functionName) {
                                $this->pluginHooks[$hook][$functionName][$dirName] = $dirName ;
                            }

                        }
                    }

                }
            }
        }
    }

    public function getRegisteredPlugins()
    {
        return $this->pluginHooks;
    }

    public static function executePluginHooks($hookName , $plugins)
    {
        $pluginPath = dirname(__FILE__).'/../plugins/';
                $returnValues = '';
                $order = 0;
        if (isset($plugins[$hookName])) {
            foreach ($plugins[$hookName] as $functionName => $pluginName) {
                foreach ($pluginName as $name) {
                    $className = $pluginPath.$name.'/'.$name.'.php';
                    include_once($className);

                    $hookPlugin = new $name();
                    $enable = $hookPlugin->loadParameters($pluginPath.$name.'/');
                    if ($enable == 'true') {
                        $order++;
                        $hookPlugin->init();
                        $output = $hookPlugin->$functionName();
                        if (isset($hookPlugin->order))
                            $order += $hookPlugin->order;

                        if ($output != '')
                            $returnValues[$order] = $output;
                    }
                }
            }
        }

        if ( is_array($returnValues))
            ksort($returnValues);

        return $returnValues;

    }
}
