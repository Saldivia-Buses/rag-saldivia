<?php
/**
* Description: Histrix Master Class
 * @date 19/11/2011
 * @author luis
 */
class Histrix {

    public function  __construct() {
        $registry =& Registry::getInstance();
        $i18n = $registry->get('i18n');
        $this->i18n  = array_map("utf8_encode", $i18n);
        $this->dateFormat = $registry->get('dateFormat');        

    }


}
?>
