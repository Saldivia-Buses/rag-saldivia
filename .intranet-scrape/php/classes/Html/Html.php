<?php
/**
 * Automatic Tag construction
 * 2009-07-02
 *
 * @category Html_Construction
 * @package  Histrix
 * @author   Luis M. Melgratti <luis@estudiogenus.com>
 * @license  http://gnu.org/copyleft/gpl.html GNU GPL v2
 * @version  1.0
 * @link     http://www.estudiogenus.com/histrix
 *
 */

//namespace histrix\classes;
/**
* Tag Manipulation Class
 * Automatic Tag construction
 * 2009-07-02
 *
* @category Html_Construction
* @package  Histrix
* @author   Luis M. Melgratti <luis@estudiogenus.com>
* @license  http://gnu.org/copyleft/gpl.html GNU GPL v2
* @version  1.0
* @link     http://www.estudiogenus.com/histrix
*/
class Html
{
    /**
    * Create Tag
    *
    * @param string $tag     Tag Name
    * @param String $content Tag content
    * @param array  $param   aditional attributes
    *
    * @return string tag string
    *
    */
    public static function Tag($tag, $content, $param ='')
    {
        $strParam = '';
        if (isset($param) && is_array($param)) {

            foreach ($param as $key => $value) {
                //if ($value != '')
                $strParam .= ' '.$key.'="'.$value.'"';
            }
        }

        $salida = '<'.$tag.$strParam.'>'.$content.'</'.$tag.'>';

        return $salida;
    }

    /**
    * Create javascript tag
    *
    * @param mixed $code String or array of javascript code lines
    *
    * @return string javascript tag string
    *
    */
    public static function scriptTag($code)
    {
        $str = $code;

        if (is_array($code)) {
            $str = implode(';', $code);
        }

        if ($str != '') {
            //$salida = '<script type="text/javascript">'.$str.'</script>';
            $salida = Html::Tag('script', $str,  array('type' => 'text/javascript'));
        }

        return $salida;
    }

    /**
     * Convert an Array to a javascript Object
     *
     * @param array $array variable names and values
     *
     * @return string javascript object
     */

    public static function javascriptObject($array , $quotes="\'")
    {
        $str = '{';
        $delimiter = '';
        foreach ($array as $key => $value) {
            if ($value != '') {
                $str .= $delimiter.' '.$quotes.$key.$quotes.':'.$value;
                $delimiter = ',';
            }
        }
        $str .= '}';

         return $str;
    }

}
