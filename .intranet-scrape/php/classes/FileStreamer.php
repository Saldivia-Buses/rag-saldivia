<?php
class FileStreamer
{
	private $_fileName;
	private $_contentLength;
	private $_destination;
	
	public function __construct()
	{
		if (!isset($_SERVER['HTTP_X_FILE_NAME']) 
			&& !isset($_SERVER['CONTENT_LENGTH'])) {
			throw new Exception("No headers found!");
		}

		$this->_fileName = rawurldecode($_SERVER['HTTP_X_FILE_NAME']);
		$this->_contentLength = $_SERVER['CONTENT_LENGTH'];
	}
    
    public function isValid()
    {
        if (($this->_contentLength > 0)) {
            return true;
        }

        return false;
    }

    public function fileName (){
        return $this->_fileName;
    }
    
    public function filePath(){
	return $this->_destination . $this->_fileName;
    }

    public function setDestination($destination)
    {
    	$this->_destination = $destination;
    }
    
    public function receive()
    {
        if (!$this->isValid()) {
        	throw new Exception('No file uploaded!');
        }
        if (!is_dir($this->_destination))
          mkdir($this->_destination, 0777, true);

        $status = file_put_contents(

        	$this->_destination . $this->_fileName, 
        	file_get_contents("php://input")
            );

//        loger($this->_destination . $this->_fileName, 'drop2');
        return true;
    }
}
?>