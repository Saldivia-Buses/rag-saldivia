<?php
/**
 * Notification Class
 *
 * @author Luis M. Melgratti
 * @creation 2009-09-19
 */

class Notification {

    public function __construct($id, $title='', $text='', $fade= 0, $class= 'notification'){
        $this->id = $id;
        $this->title = $title;
        $this->contentString = $text;
        $this->class =$class;
        $this->fade = $fade;
        $this->parameters = null;
        
    }
    // como implemento esto?
    public function onClick( $function){
        $this->click = $function;

    }

    public function buildText(){

        if (isset($this->parameters['link'])){
            $program = $this->parameters['link'];
            $xmldir  = $this->parameters['dir'];
            $text   =  $this->parameters['text'];

            $button = '<button  onclick="Histrix.loadXML( \'DIV'.$program.'\', \'histrixLoader.php?&xml='.$program.'&dir='.$xmldir.'\', \''.$text.'\')"  type="button">'.$text.'</button>';
       }
        $this->contentString = addslashes($button);
        $regexp = "/([a-zñÑA-Z0-9-. ]*): ([a-zñÑA-Z0-9-.' ]*),/";

        if (isset($this->key)){
/*            $texto = str_replace('\'', '', $this->key);
            $texto = str_replace(',', '<br>', $texto);*/
            preg_match_all($regexp, $this->key, $arr, PREG_PATTERN_ORDER);
            $terminos = count($arr[0]);
            foreach($arr[0] as $index => $terms){

                $value = Types::formatDate(str_replace("'", '', $arr[2][$index]));

                $texto .= '<tr><th>'.$arr[1][$index].'</th><td>'.$value.'</td></tr>';
            }
            
            $this->contentString .= addslashes('<table class="changes key">'.$texto.'</table>');

        }
        
        if (isset($this->obs)){
            preg_match_all($regexp, $this->obs, $arrObs, PREG_PATTERN_ORDER);
            foreach($arrObs[0] as $index => $terms){
                if ($arrObs[2][$index] != ''){

                    $value = Types::formatDate(str_replace("'", '', $arrObs[2][$index]));

                    $textoObs .= '<tr><th>'.$arrObs[1][$index].'</th><td>'.$value.'</td></tr>';
                }
            }
            $this->contentString .= addslashes('<div class="divchanges"><table class="changes key">'.$textoObs.'</table><div>');
        }

        if (isset($this->log)){
            $this->log = preg_replace("/[\n]/",'',$this->log);
            $this->contentString .= addslashes($this->log);
        }

    }

    public function createTag(){
        $this->code[] = "Histrix.notification('$this->id', {title:'$this->title', text:'$this->contentString', notifClass:'$this->class', click: function(){ $this->click }, fade:$this->fade });";
        
    }

    public function output(){
        $this->createTag();
        return Html::scriptTag($this->code);
    }
}
?>
