#include <boost/multiprecision/cpp_int.hpp>
#include <iostream>

int main() {
    boost::multiprecision::cpp_int num = 1;
    
    for(int i = 2; i <= 2017; ++i) {
        num *= i;
    }
    
    while(num % 2 == 0) {
        num /= 2;
    }
    
    std::cout << num << std::endl;
}
