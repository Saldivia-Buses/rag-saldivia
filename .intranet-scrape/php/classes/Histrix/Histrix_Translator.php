<?php
/**
* Description: Histrix Master Class
 * @date 19/11/2011
 * @author luis
 */
class Histrix_Translator {

    public function  __construct($dirXML='', $realPath='',$file='') {
        $this->realPath = $realPath;
        $this->dirXML = $dirXML;
        $this->file = $file;
        $this->defaultLang = 'es';

        if ($this->dirXML =='')
            $this->dirXML = $_SESSION['dirXML'];

        /*
        $registry =& Registry::getInstance();
        $i18n = $registry->get('i18n');
        $this->i18n  = array_map("utf8_encode", $i18n);
        */
    }

    // Read TMX file from disk
    private function readTmx (){

        $path = dirname($this->realPath);
        $c = 0;
        while ($path.'/' != $this->dirXML && $c < 5 ){
            $tmxfile = $path.'/'.basename($path).'.tmx';
            $path = dirname($path);
            $c++;

            if ($this->file != '')
                $tmxfile = dirname($this->dirXML).'/'.$this->file;

            if (is_file($tmxfile)) {

                $this->xmlfile = @simplexml_load_file($tmxfile);

                if (is_object($this->xmlfile->body->tu)) {

                    foreach ($this->xmlfile->body->tu as $tunit) {
                        $tuid = (string) $tunit['tuid'];
                        foreach ($tunit->tuv as $tuv) {
                            $lang = (string) $tuv->attributes('xml', 1)->lang;
                            // check for malformed namespace
                            if ($lang == '') {
                                $lang = (string) $tuv['lang'];
                            }
                            $this->tmx[$tuid][$lang] = (string) $tuv->seg;
                        }
                    }
                }
                $this->newtmxfile = $tmxfile;

                return true;
            }
        }

        return false;
    }

    public function translate($str_in){
        $str = (string) $str_in;

        $string = trim(strtolower($str));
        $string = html_entity_decode($string,ENT_QUOTES, 'UTF-8');

        if ($str == '') return;
        
        $curlang = $_SESSION['lang'];

        // Read file from disk
        if (!is_array($this->tmx)) {

            $filefound = $this->readTmx();

            if (!$filefound) 
                return  $str;

        }

        // if language same as adminlang do not translate
        // add Item to tmx  file

        if (!array_key_exists($string, $this->tmx) ) {

            $this->tmx[$string][$this->defaultLang] = $string;

            if (isset($this->xmlfile) && 
                is_object($this->xmlfile) && 
                is_writable($this->newtmxfile)){

                $newtu = $this->xmlfile->body->addChild('tu');

                $newtu->addAttribute('tuid', $string);
                $newtu->addAttribute('datatype', "plaintext");

                $newtuv = $newtu->addChild('tuv');
                $newtuv->addAttribute('xml:lang', $this->defaultLang );
                $newtuv->addChild('seg', html_entity_decode($str, ENT_QUOTES, 'UTF-8'));

                $dom = new DOMDocument('1.0');
                $dom->preserveWhiteSpace = false;
                $dom->formatOutput = true;
                $xmlString = $this->xmlfile->asXML();
               // loger($xmlString, 'xml');

                $dom->loadXML($xmlString );
                
                if (is_writable($this->newtmxfile))
                     file_put_contents($this->newtmxfile, $dom->saveXML());
            }
        }

        $translation = (isset($this->tmx[$string][$curlang]) &&  $this->tmx[$string][$curlang] != ''  )?$this->tmx[$string][$curlang]:$str;

        // adapt length
        if (strlen($str_in) > strlen($string)){
            $translation = str_pad($translation,strlen($str_in), ' ');
        }

        return $translation;
    }

}
?>