<?php
/* 
 * 2009-08-05
 * Widget Class
 * 
 */

class Widget{

    public function __construct($id){
        $this->id = $id;
        $this->uid = uniqid('widget');
        
    }

    public function display(){


        if ($this->width != ''){
            $style = 'width:'.$this->width.';';
        }        
        if ($this->height != ''){
            $style .= 'height:'.$this->height.';';
        }
        if ($this->style != ''){
            $style .= $this->style;
        }
        if ($style != ''){
            $style = ' style="'.$style.'"';
            
        }

        $html = '<div class="widget" '.$style.' >';

        $js = $this->loadData();

        $html .= $this->topBar($js);
        $html .= $this->Data();
        $html .= '</div>';

        if ($js != ''){
             $html .= Html::scriptTag($js);
        }
        return $html;
    }

    public function topBar(){
        $html  = '<div class="barrasup">';
        $html .= $this->title;
        $html .= '<span style="float:right;">';
        $html .= '<img id="print'.$this->uid.'" src="../img/printer1.png" width="16px"/>';
        $html .= '<img id="setup'.$this->uid.'" src="../img/setup16.png" width="16px"/>';
        $html .= '<img id="reload'.$this->uid.'" src="../img/view-refresh.png" width="16px"/>';
        $html .= '</span>';
        $html .= '</div>';
        return $html;

    }

    public function Data(){
        if ($this->style != ''){
            $style= ' style="'.$this->style.'" ';
        }
        $html = '<div id="'.$this->uid.'" '.$style.'>';
        $content = $this->text;
        $content .= $this->iframe();
        if ($content == ''){
            $content .= '<div id="throbber" ></div>';
        }
        $html .= $content;

        $html .= '</div >';
        return $html;
    }
    public function iframe(){
        if ($this->iframe != ''){
            $html .= '<iframe src="'.$this->iframe.'" width="99%"></iframe>';
        }

        return $html;
    }


    public function loadData(){
        if ($this->url != ''){
            $url = $this->url.'&dashboard=true&widgetId='.$this->uid;
            $js[] = '$("#'.$this->uid.'").load("'.$url.'");';
            $js[] = '$("#reload'.$this->uid.'").click( function(){ $("#'.$this->uid.'").html("<div id=\"throbber\" ></div>").load("'.$url.'")});';
            $js[] = '$("#setup'.$this->uid.'").click( function(){ $("fieldset", "#'.$this->uid.'").slideToggle();})';
            $js[] = '$("#print'.$this->uid.'").click( function(){ 
                var wimg  = window.open( $( "#'.$this->uid.' .chart").attr("src") ,"'. $this->title.'" );
                wimg.print();
            })';            

        }

        return $js;
    }
}

?>
