<?php
/* 
* Histrix Tray class
 * place holder for events
 *  */

/**
 * Description of Histrix_Tray
 *
 * @author luis
 */
class Histrix_Tray extends Histrix {
    //put your code here
    public function  __construct($class ='', $url='') {
        parent::__construct();
        $this->className = $class;
        $this->plugins = Cache::getCache('Plugins');
        if ($url != ''){
            $this->supportUrl = $url;
            $this->supportLink = 'histrix';
        }
        else {
            $this->supportUrl = 'http://www.estudiogenus.com/usuario-de-histrix';
            $this->supportLink = '<img src="../img/histrixico.gif" alt="histrix" width="16" title="'.$this->supportUrl.'"/>';
            $this->link = "Histrix.loadInnerXML('about','../about.php',    null, 'Acerca de Histrix', null, null, {width : '400px', height : '200px',modal : true}); return false;";
        }
	    
        $this->build();
    }

    public function build(){

        $output  = '<div class="'. $this->className .'" >';

        $output .= '<div id="histrixBrand" style="float:right;">';
        $output .= '<a class="ttf_ubuntu" style="color:#aaa; font-size:16px;outline: none;" onclick="'.$this->link.'"  href="'.$this->supportUrl.'" target="_blank">'.$this->supportLink.'</a>';
        $output .= '</div>';

        $output .= '<div id="Histrix_Tray">';
        $pluginOutput =  PluginLoader::executePluginHooks('trayInit', $this->plugins);
        if (is_array($pluginOutput)) $output .= implode($pluginOutput);
        $output .= '</div>';
        
        $output .= '</div>';
        $this->output = $output;
    }

    public function render(){
        echo $this->output;
    }



}
?>
