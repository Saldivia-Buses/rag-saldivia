<?php

/**
 * 
 *  MENUBAR CLASS
 * 19/11/2011
 */
class Histrix_Menu {

    public function __construct($type = 'menubar', $orientation ='h'){
        $this->type = $type;
        $this->orientation = $orientation;
        $this->i = 0;
        $this->j = 1;
        $this->translator = new Histrix_Translator('','','xml/histrix/menu/menu.tmx');
    }

    function render($return=false, $style=''){

        switch ($this->type){
                case 'menubar':
                    $output = '<div style="'.$style.'" class="hmenugradient">';
                    $output .= $this->menuString;
                    $output .= '</div>';

                break;
                case 'button':
                    $output = '<div>' ;
                    $output .= $this->menuString;
                    $output .= '</div>';

                break;
                case 'icons':

                break;
                case 'acordeon':

                break;
        }

        if ($return) return $output;
        else 
        	echo $output;
    }

    function build($nivel, $id = '', $perfil = 0, $jsmenu = '', $idMenu = '', $option ='') {
        $quotes = '';


        $tablaMenu = 'HTXMENU';
        $tablaclave = 'Id_menu';
        $campoejecuta = 'ejecuta';
        $orden = 'orden';
        $menuId = 'menuId';
        $campohijo = 'hijo';
        $campoparametros = 'parametros';

	if (!isset($option['login']))
	    $option['login']= '';
        $strSQL = 'select *, ' . $tablaMenu . '.menuId as "MID"  from ' . $tablaMenu;
        $strSQL .= ' left join HTXPROFILE_AUTH on ' . $quotes . 'Id_perfil' . $quotes . '=' . $perfil;
        $strSQL .= ' and ' . $quotes . 'HTXPROFILE_AUTH.menuId' . $quotes . '= HTXMENU.menuId';
        $strSQL .= ' left join HTXUSER_AUTH on ' . $quotes . 'login' . $quotes . '= "' . $option['login'] . '"';
        $strSQL .= ' and ' . $quotes . 'HTXUSER_AUTH.menu_id' . $quotes . '= HTXMENU.menuId';
        $strSQL .= ' where ' . $quotes . $tablaMenu . '.' . $tablaclave . $quotes . '=' . "'" . $nivel . "'";
        $strSQL .= ' order by ' . $quotes . $tablaMenu . '.' . $orden . $quotes . '';

        $result = consulta($strSQL, null, '_nolog');
        $cont = 0;
        if ($id != '')
            $id = 'id="' . $id . '"';


        $salida = '';
        while ($row = _fetch_array($result)) {
            $cont++;
            $modulo = '';
            $hijo = $row[$campohijo];
            $deny = $row['deny'];
            //deny some items
            if ($deny == 1) continue;
            
            $opcionesMenu = $row[$campoparametros];
            $modulo = $row[$campoejecuta];

            if ($modulo == 'AbmGenerico.php' || $modulo == 'histrixLoader.php')
                $modulo = '';

            $miorden = $row[$orden];
            $mimenuId = $row['MID'];

            $Descripcion = strtolower(htmlentities($row["titulo"], ENT_QUOTES, 'UTF-8'));

            // HIDDE dot entries IN MENU
            if (substr($Descripcion, 0,1) == '.' )
            	continue;
            $Descripcion = ucfirst($this->translator->translate($Descripcion));

            
            $extra = '';
            if (isset($row['extrafield_1'])) {
                $extra1 = $row['extrafield_1'];
            }
            if (isset($row['extrafield_2'])) {
                $extra2 = $row['extrafield_2'];
            }


            $helplink = '';

            $helplink1 = '';
            $helplink2 = '';
            
            if (isset($row['helplink_1'])) {
                $helplink1 = $row['helplink_1'];
                $helppath1 = 'hlppath="' . urlencode($helplink1) . '"';
            }
            if (isset($row['helplink_2'])) {
                $helplink2 = $row['helplink_2'];
                $helppath2 = 'hlppath="' . urlencode($helplink2) . '"';
            }

            if ($extra1 != '') {
                $Descripcion .= '  <span ' . $helppath1 . ' class="menu_extra"> ' . $extra1 . '</span>';
            }

            if ($extra2 != '') {
                $Descripcion .= '  <span ' . $helppath2 . ' class="menu_extra"> ' . $extra2 . '</span>';
            }


            $hay = false;

            if ($idMenu != '' && $this->orientation == 'h' && $this->type != 'button')
                $salida .= '<td>';

            if ($hijo != '') {
                $this->j++;
                $h = $this->j;

                $salidasubmen = $this->build($hijo, '', $perfil, $jsmenu, '', $option);

                if ($salidasubmen != '') {

                    $salida .= "\n";
                    
                    if ($nivel == "phpmen" && $this->type != 'button')
                        $salida .= '<a class="button" >' . $Descripcion . '</a>';
                    else
                        $salida .= '<a class="item" >' . $Descripcion . '<img class="arrow" width="4px" height="7px;" src="../img/arrow1.gif"/></a>';
//                           loger($Descripcion, 'menu');
                    $salida .= '<div class="section" id="' . $h . '">' . $salidasubmen . '</div>';
                    $hay = true;
                }
            } else {
                $opciones = '';
                if ($opcionesMenu != 'http' && $opcionesMenu != 'https') {
                    if (substr($opcionesMenu, 1, 1) == '&') {
                        $opciones .= '?' . $opcionesMenu;
                    } else {
                        $opciones .= '?' . $opcionesMenu;
                    }
                    $opciones = htmlentities($opciones);
                    $campos = '';

                    parse_str($opciones, $campos);
                    $ampstr = '';
                    if (isset($campos['amp;xml']))
                        $ampstr = $campos['amp;xml'];

                    $xml = 'DIV' . $ampstr;

                    if ($xml == 'DIV') {
                        $xml = md5($modulo . $opciones);
                    }
                    $xml = str_replace('.', '_', $xml);


                    $notifica = '';

                    // profile authorization
                    if ($row['Id_perfil'] != '')
                        $hay = true;

                    // user authorization
                    if ($row['login'] != '') {
                        $hay = true;
                    }

                    if ($perfil == 1 || $option['all'] == 1)
                        $hay = true;


                    if (($hay)) {

                        $sqlNot = 'select * from HTXNOTIFUSER where ' . $quotes . 'menuId' . $quotes . '=' . $menuId . ' LIMIT 1';

                        $rs2 = consulta($sqlNot, null, 'nolog');

                        $notifica = ', null';
                        while ($row = _fetch_array($rs2)) {
                            $notifica = ',' . $mimenuId;
                        }

                        $this->j++;
                        $this->i++;

                        $salida .= "\n";
                        $loader = ($modulo != '') ? ' loader="' . $modulo . '" ' : '';

                        $div = '<a class="item"  helplink="' . $helplink . '" rel="' . $opciones . '" menuId="' . $mimenuId . '"' . $loader . '>';
                        $div .= $Descripcion;
                        $div .= "</a>";
                        $salida .= $div;
                    }
                } else {
                    $this->j++;
                    $salida .= "\n";
                    $salida .= "		<a  class=\"item\" " .
                            ' onclick="Histrix.loadExternalXML(' .
                            "'contenido', '$sisdir$modulo$opciones');" .
                            '" >';

                    $salida .= $Descripcion;
                    $salida .= '</a>';
                    $hay = true;
                }
            }
            if ($idMenu != '' && $this->orientation == 'h' && $this->type != 'button')
                $salida .= '</td>';
        }

        if ($idMenu != '') {

            $style = (isset($option['style'])) ? 'style="' . $option['style'] . '"' : '';

            $salidaUL = '<table width="100%" cellspacing="0" cellpadding="0" id="' . $idMenu . '" class="XulMenu" ' . $style . ' ' . $id . '><tr>';

            if ($this->type == 'button'){
                $salidaUL .= '<td><a class="button">Histrix Menu</a><div class="section" id="0">' . $salida .'</div></td>';;
            }
            else{

                if ($this->orientation == 'v')
                    $salida = '<td>' . $salida . '</td>';

              $salidaUL .= $salida;

            }
            

            $salidaUL .= '</tr></table >';
        }
        else{
            return $salida;
        }

        $this->menuString = $salidaUL;
        return $salidaUL;
    }

}

?>