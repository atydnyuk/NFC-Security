#!/usr/bin/python

from optparse import OptionParser

def main():
    parser = OptionParser()
    parser.add_option("-f", "--file", dest="filename", default="log.txt",
                      help="Choose input file", metavar="FILE")
    
    (options, args) = parser.parse_args()
    
    f = open(options.filename,"r")
    text = f.read()
    request_count = 0
    good_count = 0
    bad_count = 0
    state = True
    dropped_count = 0.0
    for line in text.split("\n"):
        if "request" in line:
            request_count += 1
            if "Good" in line:
                if (not state):
                    state=True
                    dropped_count+=1
                good_count+=1
            if "Malicious" in line:
                if (state):
                    state=False
                bad_count+=1

    print "We recieved "+str(request_count)+" requests\n"
    print "Of these, "+str(good_count)+" were good and "+str(bad_count)+" were malicious\n"
    print "There were "+str(dropped_count)+" good requests dropped\n"
    print "This is "+str((dropped_count/good_count)*100)+"% of the total\n"
    
    
    
if __name__=="__main__":
    main()
