<?php
/**
 * Description of Histrix_Plugin
 * 2010-01-05
 *
 * @author Luis Melgratti
 */
class Histrix_Plugin extends Histrix implements PluginInterface {

    public function __construct() {
        parent::__construct();
        $this->location= dirname(__FILE__);
        
        $this->setDefaultOptions();
        $this->registerFunction('javascriptInit', 'registerJavascriptHooks');
        
    }

    public function init() {

    }

   
    public function getParameter($name){
	$parameter = isset($this->ini_array[$name])?$this->ini_array[$name]:null;
        return $parameter;
    }

    public function setDefaultOptions(){


    }
    
    public function addOption($name, $values){
    	$this->defaultOptions[$name] = $values;
    	
    }
    
    public function createOptions(){
    	$defaultOptionValues['description'] = $this->ini_array['name'];
    	$defaultOptionValues['value'] = 'false';
    	$this->addOption('enable', $defaultOptionValues);
    	
    	$options = $this->defaultOptions;
    	
   	
    	foreach($options as $name => $optionValues){
	    	$sql = 'insert ignore into HTXOPTIONS set option_description = \''.utf8_decode($optionValues['description']).'\', option_name = \'PLUGIN::'.get_class($this).'::'.$name.'\', option_value = \''.$optionValues['value'].'\'';
	    	$rs = @consulta($sql); // non bloking insert
    	}
    	 
    }
    
    public function loadParameters($path='') {
        // Parse with sections
        if ($path == ''){
            $path = dirname(__FILE__);
        }
        $file = $path.'/'.get_class($this).'.ini';
        if (file_exists($file)) {
            $this->ini_array = @parse_ini_file($file, true);
        }

    	$enable = -1;
        $sql = 'select option_value, option_name from HTXOPTIONS where option_name like \'PLUGIN::'.get_class($this).'::%\' and login=""';

        $rs = consulta($sql);
        
        if (is_object($rs)){
	        while ($row = _fetch_row($rs)){
		        $val 	= $row[0];
		        $start  = strlen('PLUGIN::'.get_class($this).'::');
		        $name   = trim(substr($row[1], $start));
		        
		        $this->ini_array[$name]=$val;

		        $enable = $this->ini_array['enable'];
		        
	        }
        }
        
        //load Lang Files
        $langfile = $path.'/lang/'.$_SESSION['lang'].'.php';
        if (file_exists($langfile)){
        	include($langfile);
        	$this->i18n = array_merge($this->i18n, $i18n);

        }
        return $enable;
    }

    
    function registerFunction($hookName, $functionName) {
        $this->functionHooksArray[$hookName][$functionName] = $functionName;
    }

    function registerJavascriptFunction($hookName, $functionName) {
        $this->javascriptHooksArray[$hookName] = $functionName;
    }

    /** Returns an Array of functions calls to named Hooks
     *
     * @return Array  Plugin[HookName][functionName]
     */
    public function getFunctionHooks() {

        return $this->functionHooksArray;
    }

    public function registerJavascriptHooks(){
        $jsHooks = '';
        if (is_array($this->javascriptHooksArray)){
            foreach($this->javascriptHooksArray as $hookName => $functionName){
                $jsHooks .= "Histrix.registerHook('$hookName', $functionName);";
            }
        }
        return $jsHooks;
    }
}
?>
