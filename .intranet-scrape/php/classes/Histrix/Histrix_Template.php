<?php
/**
 * 
 * Histrx Template based printing
 * 2009-04-14 - Luis M. Melgratti
 * 
 *
 */

class Histrix_Template extends Histrix_txt {

    function __construct($template_file) {
        $this->tpl_file = $template_file;
    }
    
    function asignData($vars) {
    // $vars = utf8_decode($vars);
        $this->vars= (empty($this->vars)) ? $vars : $this->vars . $vars;
    }

    function format($format, $var) {
        if ($var != '')
            $var = sprintf($format , $var );
        return $var;
    }

    function parse() {
        if (!($this->fd = @fopen($this->tpl_file, 'r'))) {
        // sostenedor_error('error al abrir la plantilla ' . $this->tpl_file);
        } else {
            $this->template_file = fread($this->fd, filesize($this->tpl_file));
            fclose($this->fd);
            $this->miTXT = $this->template_file;

            $this->miTXT = str_replace ("'", "\'", $this->miTXT); // escape quotes
            $pattern[]= '/\{([a-zA-Z0-9\-_\[\]\"]*?)\}\[\[(%[a-z0-9\.\-_\[\]\"]*?)\]\]/';
            $pattern[]= '#\{([a-z0-9A-Z\-_\[\]\"]*?)\}#is';
            $replace[]= '\'.$this->format("\\2",$\\1).\'';
            $replace[]= "'.$\\1.'";
            $this->miTXT = preg_replace($pattern, $replace, $this->miTXT);


            reset ($this->vars);
            while (list($key, $val) = each($this->vars)) {
                $$key = $val;
            }
            eval("\$this->miTXT = '$this->miTXT';");
            reset ($this->vars);
            while (list($key, $val) = each($this->vars)) {
                unset($$key);
            }
            $this->miTXT=str_replace ("\'", "'", $this->miTXT);
        // echo '<pre>'.$this->miTXT.'</pre><hr>';
        //
        //  die();

        }
    }
    public function Output($fileName='', $type='') {

        if ($fileName == '') {
            return $this->miTXT;
        }

        file_put_contents($fileName, $this->miTXT);
    }




}
?>