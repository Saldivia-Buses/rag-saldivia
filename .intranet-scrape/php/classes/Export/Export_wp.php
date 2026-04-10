<?php
/* 
 * Loger Class 2013-04-28
 * Export class
 * @author Luis M. Melgratti
 */

class Export_wp extends Export{
    

    public function prepareFile(){
        
        require "../lib/IXR_Library/IXR_Library.php";
        // connect to wordpress
        /*
        $this->wpdomain = 'http://saldiviabuses.com.ar';
        $this->wpUSER = 'histrix';
        $this->wpPASS = 'genus2041';
        */

        $this->client = new IXR_Client($this->wpdomain.'/xmlrpc.php');
     //   $this->deletePreviousPosts();

    }


    public function deletePreviousPosts(){
        $params = array(0,$this->wpUSER,$this->wpPASS, 20); // Last Parameter tells how many posts to fetch
 

         // DELETE PREVIOUS POSTS
         // Run a query To Read Posts From Wordpress
        if (!$this->client->query('metaWeblog.getRecentPosts', $params)) {
          die('Something went wrong - '.$this->client->getErrorCode().' : '.$this->client->getErrorMessage());
        }
         
        $myresponse = $this->client->getResponse(); 
        $i=0; 

        foreach ($myresponse as $res) {
            //print_r($res);

            if (isset($res['custom_fields']) && is_array($res['custom_fields'])){
                foreach ($res['custom_fields'] as $num => $field) {
                    if ($field['key'] == 'histrix_xml' && 
                        $field['value'] == $this->Container->xml){

                        // Delete previous POST
                        $this->client->query('metaWeblog.deletePost', '' , $res['postid'] ,$this->wpUSER,$this->wpPASS );

                    }
                }

            }


        }
    }

    public function header(){

    }

    public function footer(){

    }

    public function processData($row, $Field, $params){

        $valor = $params['value'];

        if (count($Field->opcion) > 0 && $Field->TipoDato != "check" && $Field->valop != 'true') {
            $valor = $Field->opcion[$valor];
            if (is_array($valor)) $valor = current($valor);
        }

        $props = get_object_vars($Field);
        foreach($props as $prop => $defaultValue) {
            
            if (substr($prop, 0, 3) == 'wp_'){
                $wpvalue = $valor;

                switch ($Field->TipoDato) {
                    case "numeric" :
                        $wpvalue = $valor;
                    break;
                    default:
                        $wpvalue = utf8_decode($valor);
                    break;
                }            

                $key = substr($prop, 3);
                if ($key == 'categories'){
                    $this->content[$key] = array($wpvalue); 
                } else {
                    $this->content[$key] = $wpvalue; 
                }
                

            }
      
        }

        $this->content['publish'] = true;
        $this->content['custom_fields'] = array(array('key'=>'histrix_xml', 'value'=> $this->Container->xml ));

    }

    public function endRow(){

    }

    public function sendHeaders(){
   

    }

    public function out(){

        // var_dump($this->content);
        if (!$this->client->query('metaWeblog.newPost','', $this->wpUSER,$this->wpPASS, $this->content, true)) {
            var_dump($client->getResponse());
            die( 'Error while creating a new post' . $client->getErrorCode() ." : ". $client->getErrorMessage());  
        }
        $ID =  $this->client->getResponse();
       
        if($ID)
        {
            echo 'Post published with ID:#'.$ID;
        }        
    }


}
?>
