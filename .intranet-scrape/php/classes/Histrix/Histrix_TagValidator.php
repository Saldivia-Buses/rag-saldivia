<?php

/**
 *Histrix custom tag fixer
 * Object otoolriented version of old leoxml.php
 * Luis M. Melgratti
 */
class Histrix_TagValidator {

    /**
     * Constructor
     * @param string $dirXML
     * @param string $xml
     * @param bool $sub
     * @param string $xmlOrig
     * @param bool $subdir
     * @param bool objEmbebido
     */
    var $parameters;
    var $FilterContainer;
    var $Filter;

    function __construct() {

        $path = '../database/'.$_SESSION['datapath'];
        $this->tagfile = $path.'tags.xml';

    }

    /**
     * fix OLD tags Method from $file
     */
    function fixTags($file) {
        //echo $file;
        if (is_file($file)) {

            $tagTranslation['abm']      = 'histrix';
            $tagTranslation['consulta'] = 'form';
            $tagTranslation['tabla']    = 'table';
            $tagTranslation['titulo']   = 'title';
            $tagTranslation['IF']       = 'if';
            $tagTranslation['orden']    = 'order';
            $tagTranslation['filtros']  = 'filters';
            $tagTranslation['verdadero']= 'true';
            $tagTranslation['falso']    = 'false';
            $tagTranslation['opciones'] = 'options';
            $tagTranslation['opcion']   = 'option';            
            $tagTranslation['formato']   = 'format';           
            $tagTranslation['expresion']   = 'expression';                           


            // transform array
            foreach ($tagTranslation as $sourceTag => $targetTag) {
                // open with parameters
                $tagOrig[] = '<'.$sourceTag.' '; 
                $tagDest[] = '<'.$targetTag.' ';
                // open without parameters
                $tagOrig[] = '<'.$sourceTag.'>';
                $tagDest[] = '<'.$targetTag.'>';
                // close tag
                $tagOrig[] = '</'.$sourceTag.'>';
                $tagDest[] = '</'.$targetTag.'>';

            }

            $count = 0;
            $content = file_get_contents($file);
            $content = str_replace($tagOrig, $tagDest, $content, $count);

            if ($count > 0) {

                $f = @fopen($file, 'w');
                if ($f) {
                    loger($file, 'replace.log');
                    fwrite($f, $content);
                    fclose($f);
                } else {
                    loger('failed to update file' . $file, 'replace_error.log');
                    //   die('failed to update file'.$file);
                }
            }
        }
    }

    /**
    *This will try to help fixing attributes
    */


    function readTagFile(){

    // Read tag  translation file file from disk


            $tagfile = $this->tagfile;

            if (is_file($tagfile)){
                $this->tagxmlfile = simplexml_load_file($tagfile);
                foreach ($this->tagxmlfile->tags->tag as $tag) {
                    $tagname = (string) $tag['name'];
                    foreach ($tag->attribute as $attribute){
                        $attribute = (string) $attribute;
                        $this->tags[$tagname][$attribute] = $attribute;

                    }
                }

                $filefound = true;
                
            }
            else {
                // create tag file if not exists

                $dom = new DOMDocument('1.0');
                $dom->preserveWhiteSpace = false;
                $dom->formatOutput = true;

                $rootElt    = $dom->createElement('histrixTags');
                $rootNode   = $dom->appendChild($rootElt);
                $tags       = $dom->createElement('tags');

                $rootNode->appendChild($tags);

                $write = @file_put_contents($tagfile, $dom->saveXML());
                if ($write)
                    $this->tagxmlfile = simplexml_load_file($tagfile);
            }

          return $filefound;
    
    }

    function fixAttributes($tag, $attribute){
 
        //ignore comments
        if ($attribute[0] == '_') return;


        $tagfile = $this->tagfile;

        // add Item to tags  file
        if (!isset($this->tags[$tag][$attribute])){

            if (!isset($this->tags[$tag])){
                if (isset($this->tagxmlfile) && is_object($this->tagxmlfile->tags)){
                    $newtag = $this->tagxmlfile->tags->addChild('tag');
                    $newtag->addAttribute('name', $tag);
                }
            }
            else {
                
                $newtag = current($this->tagxmlfile->xpath("//*[@name='".$tag."']"));

            }
            if (isset($newtag) && is_object($newtag)){
                $newtag->addChild('attribute', $attribute);


                $dom = new DOMDocument('1.0');
                $dom->preserveWhiteSpace = false;
                $dom->formatOutput = true;
                $xmlString = $this->tagxmlfile->asXML();
        
                $dom->loadXML($xmlString );
                



                if (is_writable($tagfile))
                     file_put_contents($tagfile, $dom->saveXML());

                $this->tags[$tag][$attribute] = $attribute;          
            }
        }

    }
}

?>
