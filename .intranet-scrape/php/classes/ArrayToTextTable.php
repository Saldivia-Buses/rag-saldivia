<?php
/**
 * Array to Text Table Generation Class
 *
 * @author Tony Landis <tony@tonylandis.com>
 * @link http://www.tonylandis.com/
 * @copyright Copyright (C) 2006-2009 Tony Landis
 * @license http://www.opensource.org/licenses/bsd-license.php
 */
class ArrayToTextTable
{
    /** 
     * @var array The array for processing
     */
    private $rows;

    /** 
     * @var int The column width settings
     */
    private $cs = array();

    /**
     * @var int The Row lines settings
     */
    private $rs = array();

    /**
     * @var int The Column index of keys
     */
    private $keys = array();

    /**
     * @var int Max Column Height (returns)
     */
    private $mH = 2;

    /**
     * @var int Max Row Width (chars)
     */
    private $mW = 30;

    private $head  = false;
    private $pcen  = "+";
    private $prow  = "-";
    private $pcol  = "|";
    private $innerpcol  = " ";
    
    
    /** Prepare array into textual format
     *
     * @param array $rows The input array
     * @param bool $head Show heading
     * @param int $maxWidth Max Column Height (returns)
     * @param int $maxHeight Max Row Width (chars)
     */
    public function ArrayToTextTable($rows)
    {
        $this->rows =& $rows;
        $this->cs=array();
        $this->rs=array();
 
        if(!$xc = count($this->rows)) return false; 
        $this->keys = array_keys($this->rows[0]);
        $columns = count($this->keys);
        
        for($x=0; $x<$xc; $x++)
            for($y=0; $y<$columns; $y++)    
                $this->setMax($x, $y, $this->rows[$x][$this->keys[$y]]);
    }

    public function setTotals($hasTotals){
        $this->hasTotals = $hasTotals;
    }

    public function setCenterCharacter($char = '+'){
        $this->pcen  = $char;
    }
    public function setInnerBorder($char){
        $this->innerpcol = $char;
    }
    /**
     * Set Border Character (horizontal and Vertical
     * @param string $vert
     * @param string $horiz
     */
    public function setOuterBorder($vert, $horiz ="-"){
        $this->pcol = $vert;
        $this->prow  = $horiz;
    }

    /**
     * Show the headers using the key values of the array for the titles
     * 
     * @param bool $bool
     */
    public function showHeaders($bool)
    {
       if($bool) $this->setHeading(); 
    } 
    
    /**
     * Set the maximum width (number of characters) per column before truncating
     * 
     * @param int $maxWidth
     */
    public function setMaxWidth($maxWidth)
    {
        $this->mW = (int) $maxWidth;
    }
    
    /**
     * Set the maximum height (number of lines) per row before truncating
     * 
     * @param int $maxHeight
     */
    public function setMaxHeight($maxHeight)
    {
        $this->mH = (int) $maxHeight;
    }
    
    /**
     * Prints the data to a text table
     *
     * @param bool $return Set to 'true' to return text rather than printing
     * @return mixed
     */
    public function render($return=false)
    {
        if($return) ob_start(null, 0, true); 
  
        $this->printLine();
        $this->printHeading();
        
        $rc = count($this->rows);
        for($i=0; $i<$rc; $i++){
            if ($this->hasTotals && $i + 1 == $rc){
                    $this->printLine();
            }

            $this->printRow($i);
        }
        
        $this->printLine(false);

        if($return) {
            $contents = ob_get_contents();
            ob_end_clean();
            return $contents;
        }
    }

    private function setHeading()
    {
        $data = array();  
        foreach($this->keys as $colKey => $value)
        { 
            $this->setMax(false, $colKey, $value);
            $data[$colKey] = strtoupper($value);
        }
        if(!is_array($data)) return false;
        $this->head = $data;
    }

    private function printLine($nl=true)
    {
        print $this->pcen;
        $numRows =  count($this->cs);
        foreach($this->cs as $key => $val){
            print
                $this->prow .
                str_pad('', $val, $this->prow, STR_PAD_RIGHT);

                if ($key + 1 < $numRows ){
                    print $this->prow;
                    print $this->pcen;
                }

        }
        if($nl){
            print "\n";
            $this->line++;
        }

    }

    private function printHeading()
    {
        if(!is_array($this->head)) return false;

        print $this->pcol;
        $numRows = count($this->cs);
        foreach($this->cs as $key => $val){
            $colKey++;
            print ' '.
                str_pad($this->head[$key], $val, ' ', STR_PAD_BOTH);
                
                if ($colKey  < $numRows ){
                    print $this->innerpcol;
                }
        }
        
        print $this->pcol;
        print "\n";
        $this->line++;
        //        $renderer->setCenterCharacter( '');
        $this->printLine();
    }

    private function printRow($rowKey)
    {
        // loop through each line
        for($line=1; $line <= $this->rs[$rowKey]; $line++)
        {

            print $this->pcol;
            $numRows = count($this->keys);
            for($colKey=0; $colKey < $numRows; $colKey++)
            {
                $align = STR_PAD_RIGHT;
                $value = $this->rows[$rowKey][$this->keys[$colKey]];
                $value = filter_var($value, FILTER_SANITIZE_STRING, FILTER_FLAG_STRIP_HIGH);
                //$BoolResult = if(trim($value, '+-.,0123456789')=='');

                if (trim($value, '+-.,0123456789')=='') $align = STR_PAD_LEFT;
                
                print "";
                print str_pad(substr($value, ($this->mW * ($line-1)), $this->mW), $this->cs[$colKey], ' ', $align);
                
                if ($colKey + 1 < $numRows ){
                    print " ";
                    print $this->innerpcol;
                }
            }
            print " ";
            print $this->pcol;
            print  "\n";
            $this->line++;
        }
    }

    private function setMax($rowKey, $colKey, &$colVal)
    { 
        $w = mb_strlen($colVal);
        $h = 1;
        if($w > $this->mW)
        {
            $h = ceil($w % $this->mW);
            if($h > $this->mH) $h=$this->mH;
            $w = $this->mW;
        }
 
        if(!isset($this->cs[$colKey]) || $this->cs[$colKey] < $w)
            $this->cs[$colKey] = $w;

        if($rowKey !== false && (!isset($this->rs[$rowKey]) || $this->rs[$rowKey] < $h))
            $this->rs[$rowKey] = $h;
    }
}
?>