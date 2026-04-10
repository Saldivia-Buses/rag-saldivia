<?php
/* 
 * 2009-08-05
 * Dashborad Class
 * meta conteiner of data containers.
 */


class UI_dashboard{

    public function __construct(){
        $this->tipo = 'dashboard';
    }


    /**
     * Add Widget to Dashboard
     */
    public function addWidget(Widget $widget){
        $id  = $widget->id;
        $this->widgets[$id]=$widget;
    }

    /**
     * remove Widget to Dashboard
     */
    public function removeWidget(){

    }

    public function display(){
        $html .= '<div class="dashboard">';
        foreach($this->widgets as $id  => $widget){
            $html .= $widget->display();
        }
        $html .= '</div>';
        return $html;
    }

}

?>
