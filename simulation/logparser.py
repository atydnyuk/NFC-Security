#!/bin/bash

def main():
    parser = OptionParser()
    parser.add_option("-f", "--file", dest="filename", default="log.txt",
                      help="Choose input file", metavar="FILE")
    
    (options, args) = parser.parse_args()

    f = open(options.filename,"r")
    text = f.read().replace("\n", "")
    
    print text
    
    

if __name_=="__main__":
    main()
