<?php
/* 
 * Loger Class 2013-04-28
 * Export xml class
 * @author Luis M. Melgratti
 */

class Export_xml extends Export{
    

    public function out(){
        $mixml = new Cont2XML($this->Container);
        $mixml->exportData();

        $dest = ($dest != '')? $dest : $_GET['dest'];

        if ($dest == '') $dest = 'I';
        
        $mixml->out($dest, $this->filename, $this->filename);

    }


}
?>
